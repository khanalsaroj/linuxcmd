package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type uniqCommand struct{}

func (uniqCommand) Name() string    { return "uniq" }
func (uniqCommand) Summary() string { return "report or filter out repeated adjacent lines" }

var uniqSpec = parser.Spec{
	{Short: 'c', Long: "count"},
	{Short: 'd', Long: "repeated"},
	{Short: 'u', Long: "unique"},
}

func (uniqCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, uniqSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "uniq: %s\n", err)
		return command.ExitUsage
	}

	showCount := res.Bool('c', "count")
	onlyRepeated := res.Bool('d', "repeated")
	onlyUnique := res.Bool('u', "unique")

	var in io.Reader = ctx.Stdin
	files := paths.ExpandGlobs(res.Positional)
	if len(files) > 0 {
		arg := files[0]
		if arg != "-" {
			resolved, err := paths.Resolve(arg)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "uniq", arg, err)
				return command.ExitFailure
			}
			f, err := os.Open(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "uniq", arg, err)
				return command.ExitFailure
			}
			defer f.Close()
			in = f
		}
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var prev string
	count := 0
	first := true

	emit := func() {
		if first {
			return
		}
		if onlyRepeated && count < 2 {
			return
		}
		if onlyUnique && count > 1 {
			return
		}
		if showCount {
			fmt.Fprintf(ctx.Stdout, "%7d %s\n", count, prev)
		} else {
			fmt.Fprintln(ctx.Stdout, prev)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if first {
			prev = line
			count = 1
			first = false
			continue
		}
		if line == prev {
			count++
			continue
		}
		emit()
		prev = line
		count = 1
	}
	emit()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(ctx.Stderr, "uniq: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(uniqCommand{}) }
