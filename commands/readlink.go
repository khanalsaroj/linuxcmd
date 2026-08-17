package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type readlinkCommand struct{}

func (readlinkCommand) Name() string    { return "readlink" }
func (readlinkCommand) Summary() string { return "print a symbolic link's target" }

var readlinkSpec = parser.Spec{
	{Short: 'f', Long: "canonicalize"},
}

func (readlinkCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, readlinkSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "readlink: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: readlink [-f] LINK...")
		return command.ExitUsage
	}
	canonicalize := res.Bool('f', "canonicalize")

	exit := command.ExitSuccess
	for _, arg := range res.Positional {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "readlink", arg, err)
			exit = command.ExitFailure
			continue
		}

		if canonicalize {
			canonical, err := filepath.EvalSymlinks(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "readlink", arg, err)
				exit = command.ExitFailure
				continue
			}
			fmt.Fprintln(ctx.Stdout, canonical)
			continue
		}

		target, err := os.Readlink(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "readlink", arg, err)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintln(ctx.Stdout, target)
	}
	return exit
}

func init() { command.Register(readlinkCommand{}) }
