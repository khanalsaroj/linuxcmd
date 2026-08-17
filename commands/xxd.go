package commands

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type xxdCommand struct{}

func (xxdCommand) Name() string    { return "xxd" }
func (xxdCommand) Summary() string { return "hex dump (or reverse) a file" }

var xxdSpec = parser.Spec{
	{Short: 'g', HasArg: true},
	{Short: 'r', Long: "revert"},
}

func (xxdCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, xxdSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "xxd: %s\n", err)
		return command.ExitUsage
	}
	group := 2
	if v, ok := res.Value('g', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "xxd: invalid group size '%s'\n", v)
			return command.ExitUsage
		}
		group = n
	}
	revert := res.Bool('r', "revert")

	var in io.Reader = ctx.Stdin
	files := paths.ExpandGlobs(res.Positional)
	if len(files) > 0 && files[0] != "-" {
		resolved, err := paths.Resolve(files[0])
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "xxd", files[0], err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "xxd", files[0], err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "xxd: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	if revert {
		decoded, err := decodeXxd(string(data))
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "xxd: %s\n", err)
			return command.ExitFailure
		}
		ctx.Stdout.Write(decoded)
		return command.ExitSuccess
	}

	const perLine = 16
	for off := 0; off < len(data); off += perLine {
		end := off + perLine
		if end > len(data) {
			end = len(data)
		}
		chunk := data[off:end]
		fmt.Fprintf(ctx.Stdout, "%08x: ", off)
		for i := 0; i < perLine; i += group {
			ge := i + group
			if ge > len(chunk) {
				ge = len(chunk)
			}
			if i < len(chunk) {
				fmt.Fprint(ctx.Stdout, hex.EncodeToString(chunk[i:ge]))
			} else {
				fmt.Fprint(ctx.Stdout, strings.Repeat("  ", group))
			}
			fmt.Fprint(ctx.Stdout, " ")
		}
		fmt.Fprint(ctx.Stdout, " ")
		for _, b := range chunk {
			if b >= 0x20 && b < 0x7f {
				fmt.Fprintf(ctx.Stdout, "%c", b)
			} else {
				fmt.Fprint(ctx.Stdout, ".")
			}
		}
		fmt.Fprintln(ctx.Stdout)
	}
	return command.ExitSuccess
}

// decodeXxd rebuilds bytes from xxd's default hex-dump format: an offset,
// colon, hex bytes, then an ASCII gutter that this parser ignores.
func decodeXxd(text string) ([]byte, error) {
	var out []byte
	for _, line := range strings.Split(text, "\n") {
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		rest := line[colon+1:]
		// Stop at the ASCII gutter (two or more spaces after hex pairs).
		if gutter := strings.Index(rest, "  "); gutter >= 0 {
			rest = rest[:gutter]
		}
		hexStr := strings.ReplaceAll(strings.TrimSpace(rest), " ", "")
		bytes, err := hex.DecodeString(hexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hex data")
		}
		out = append(out, bytes...)
	}
	return out, nil
}

func init() { command.Register(xxdCommand{}) }
