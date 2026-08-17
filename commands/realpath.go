package commands

import (
	"fmt"
	"path/filepath"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type realpathCommand struct{}

func (realpathCommand) Name() string    { return "realpath" }
func (realpathCommand) Summary() string { return "print a resolved absolute path" }

func (realpathCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: realpath PATH...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range ctx.Args {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "realpath", arg, err)
			exit = command.ExitFailure
			continue
		}
		if canonical, err := filepath.EvalSymlinks(resolved); err == nil {
			resolved = canonical
		}
		fmt.Fprintln(ctx.Stdout, resolved)
	}
	return exit
}

func init() { command.Register(realpathCommand{}) }
