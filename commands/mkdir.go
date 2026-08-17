package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type mkdirCommand struct{}

func (mkdirCommand) Name() string    { return "mkdir" }
func (mkdirCommand) Summary() string { return "create directories" }

var mkdirSpec = parser.Spec{
	{Short: 'p', Long: "parents"},
	{Short: 'v', Long: "verbose"},
}

func (mkdirCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, mkdirSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mkdir: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "mkdir: missing operand")
		return command.ExitUsage
	}

	parents := res.Bool('p', "parents")
	verbose := res.Bool('v', "verbose")

	exit := command.ExitSuccess
	for _, arg := range res.Positional {
		target, err := paths.Resolve(arg)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "mkdir: cannot create directory '%s': %s\n", arg, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}

		if parents {
			err = os.MkdirAll(target, 0755)
		} else {
			err = os.Mkdir(target, 0755)
		}
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "mkdir: cannot create directory '%s': %s\n", arg, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintf(ctx.Stdout, "mkdir: created directory '%s'\n", arg)
		}
	}
	return exit
}

func init() { command.Register(mkdirCommand{}) }
