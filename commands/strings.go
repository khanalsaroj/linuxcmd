package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type stringsCommand struct{}

func (stringsCommand) Name() string    { return "strings" }
func (stringsCommand) Summary() string { return "extract printable sequences from a file" }

var stringsSpec = parser.Spec{
	{Short: 'n', HasArg: true},
}

func extractStrings(data []byte, minLen int) []string {
	var out []string
	var cur []byte
	flush := func() {
		if len(cur) >= minLen {
			out = append(out, string(cur))
		}
		cur = cur[:0]
	}
	for _, b := range data {
		if b >= 0x20 && b < 0x7f {
			cur = append(cur, b)
			continue
		}
		flush()
	}
	flush()
	return out
}

func (stringsCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, stringsSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "strings: %s\n", err)
		return command.ExitUsage
	}
	minLen := 4
	if v, ok := res.Value('n', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "strings: invalid minimum length '%s'\n", v)
			return command.ExitUsage
		}
		minLen = n
	}

	print := func(r io.Reader) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		for _, s := range extractStrings(data, minLen) {
			fmt.Fprintln(ctx.Stdout, s)
		}
		return nil
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		if err := print(ctx.Stdin); err != nil {
			fmt.Fprintf(ctx.Stderr, "strings: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "strings", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "strings", arg, err)
			exit = command.ExitFailure
			continue
		}
		err = print(f)
		f.Close()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "strings", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(stringsCommand{}) }
