package commands

import (
	"fmt"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type truncateCommand struct{}

func (truncateCommand) Name() string    { return "truncate" }
func (truncateCommand) Summary() string { return "shrink or extend a file to a given size" }

var truncateSpec = parser.Spec{
	{Short: 's', HasArg: true},
}

func (truncateCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, truncateSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "truncate: %s\n", err)
		return command.ExitUsage
	}
	sizeStr, ok := res.Value('s', "")
	if !ok {
		fmt.Fprintln(ctx.Stderr, "usage: truncate -s SIZE FILE...")
		return command.ExitUsage
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size < 0 {
		fmt.Fprintf(ctx.Stderr, "truncate: invalid size '%s'\n", sizeStr)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: truncate -s SIZE FILE...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(res.Positional) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "truncate", arg, err)
			exit = command.ExitFailure
			continue
		}
		if _, statErr := os.Stat(resolved); os.IsNotExist(statErr) {
			f, err := os.Create(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "truncate", arg, err)
				exit = command.ExitFailure
				continue
			}
			f.Close()
		}
		if err := os.Truncate(resolved, size); err != nil {
			output.SimpleErrorf(ctx.Stderr, "truncate", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(truncateCommand{}) }
