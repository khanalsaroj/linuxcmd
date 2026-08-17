package commands

import (
	"bufio"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type nlCommand struct{}

func (nlCommand) Name() string    { return "nl" }
func (nlCommand) Summary() string { return "number lines of a file" }

var nlSpec = parser.Spec{
	{Short: 'b', HasArg: true},
}

func (nlCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, nlSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "nl: %s\n", err)
		return command.ExitUsage
	}
	// -b a numbers all lines (default); -b t numbers only non-blank lines.
	numberAll := true
	if v, ok := res.Value('b', ""); ok && v == "t" {
		numberAll = false
	}

	var in *os.File
	files := paths.ExpandGlobs(res.Positional)
	if len(files) > 0 && files[0] != "-" {
		resolved, err := paths.Resolve(files[0])
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "nl", files[0], err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "nl", files[0], err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	var scanner *bufio.Scanner
	if in != nil {
		scanner = bufio.NewScanner(in)
	} else {
		scanner = bufio.NewScanner(ctx.Stdin)
	}
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	n := 1
	for scanner.Scan() {
		line := scanner.Text()
		if !numberAll && line == "" {
			fmt.Fprintln(ctx.Stdout)
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%6d\t%s\n", n, line)
		n++
	}
	return command.ExitSuccess
}

func init() { command.Register(nlCommand{}) }
