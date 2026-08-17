package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type splitCommand struct{}

func (splitCommand) Name() string    { return "split" }
func (splitCommand) Summary() string { return "split a file into pieces" }

var splitSpec = parser.Spec{
	{Short: 'l', HasArg: true},
	{Short: 'b', HasArg: true},
}

func splitSuffix(n int) string {
	// aa, ab, ac, ... az, ba, ... matching GNU split's default suffixes.
	first := byte('a' + (n/26)%26)
	second := byte('a' + n%26)
	return string([]byte{first, second})
}

func (splitCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, splitSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "split: %s\n", err)
		return command.ExitUsage
	}

	linesPer := 0
	bytesPer := 0
	if v, ok := res.Value('l', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "split: invalid line count '%s'\n", v)
			return command.ExitUsage
		}
		linesPer = n
	}
	if v, ok := res.Value('b', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "split: invalid byte count '%s'\n", v)
			return command.ExitUsage
		}
		bytesPer = n
	}
	if linesPer == 0 && bytesPer == 0 {
		linesPer = 1000
	}

	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: split [-l N | -b N] FILE [PREFIX]")
		return command.ExitUsage
	}
	prefix := "x"
	if len(res.Positional) >= 2 {
		prefix = res.Positional[1]
	}

	var in io.Reader = ctx.Stdin
	if res.Positional[0] != "-" {
		resolved, err := paths.Resolve(res.Positional[0])
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "split", res.Positional[0], err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "split", res.Positional[0], err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	partNum := 0
	newPart := func() (*os.File, error) {
		name := prefix + splitSuffix(partNum)
		partNum++
		resolved, err := paths.Resolve(name)
		if err != nil {
			return nil, err
		}
		return os.Create(resolved)
	}

	if bytesPer > 0 {
		buf := make([]byte, bytesPer)
		for {
			n, readErr := io.ReadFull(in, buf)
			if n > 0 {
				out, err := newPart()
				if err != nil {
					output.SimpleErrorf(ctx.Stderr, "split", res.Positional[0], err)
					return command.ExitFailure
				}
				out.Write(buf[:n])
				out.Close()
			}
			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				break
			}
			if readErr != nil {
				fmt.Fprintf(ctx.Stderr, "split: %s\n", output.LinuxErrorText(readErr))
				return command.ExitFailure
			}
		}
		return command.ExitSuccess
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var out *os.File
	count := 0
	for scanner.Scan() {
		if out == nil {
			out, err = newPart()
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "split", res.Positional[0], err)
				return command.ExitFailure
			}
		}
		fmt.Fprintln(out, scanner.Text())
		count++
		if count == linesPer {
			out.Close()
			out = nil
			count = 0
		}
	}
	if out != nil {
		out.Close()
	}
	return command.ExitSuccess
}

func init() { command.Register(splitCommand{}) }
