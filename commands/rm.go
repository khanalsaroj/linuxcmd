package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/fsutil"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type rmCommand struct{}

func (rmCommand) Name() string    { return "rm" }
func (rmCommand) Summary() string { return "remove files and directories" }

var rmSpec = parser.Spec{
	{Short: 'r'},
	{Short: 'R'},
	{Short: 'f', Long: "force"},
	{Short: 'v', Long: "verbose"},
}

func (rmCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, rmSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "rm: %s\n", err)
		return command.ExitUsage
	}

	force := res.Bool('f', "force")

	if len(res.Positional) == 0 {
		if force {
			return command.ExitSuccess
		}
		fmt.Fprintln(ctx.Stderr, "rm: missing operand")
		return command.ExitUsage
	}

	recursive := res.Bool('r', "") || res.Bool('R', "")
	verbose := res.Bool('v', "verbose")

	exit := command.ExitSuccess
	for _, target := range paths.ExpandGlobs(res.Positional) {
		resolved, err := paths.Resolve(target)
		if err != nil {
			if force {
				continue
			}
			output.Errorf(ctx.Stderr, "rm", "cannot remove", target, err)
			exit = command.ExitFailure
			continue
		}

		info, err := os.Lstat(resolved)
		if err != nil {
			if force {
				continue
			}
			output.Errorf(ctx.Stderr, "rm", "cannot remove", target, err)
			exit = command.ExitFailure
			continue
		}

		if info.IsDir() && !recursive {
			fmt.Fprintf(ctx.Stderr, "rm: cannot remove '%s': Is a directory\n", target)
			exit = command.ExitFailure
			continue
		}

		var removeErr error
		if info.IsDir() {
			removeErr = fsutil.RemoveRecursive(resolved)
		} else {
			removeErr = os.Remove(resolved)
		}
		if removeErr != nil {
			if force {
				continue
			}
			output.Errorf(ctx.Stderr, "rm", "cannot remove", target, removeErr)
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintf(ctx.Stdout, "removed '%s'\n", target)
		}
	}
	return exit
}

func init() { command.Register(rmCommand{}) }
