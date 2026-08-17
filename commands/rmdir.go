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

type rmdirCommand struct{}

func (rmdirCommand) Name() string    { return "rmdir" }
func (rmdirCommand) Summary() string { return "remove empty directories" }

var rmdirSpec = parser.Spec{
	{Short: 'p', Long: "parents"},
	{Short: 'v', Long: "verbose"},
}

func (rmdirCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, rmdirSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "rmdir: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "rmdir: missing operand")
		return command.ExitUsage
	}

	parents := res.Bool('p', "parents")
	verbose := res.Bool('v', "verbose")

	exit := command.ExitSuccess
	for _, arg := range res.Positional {
		target, err := paths.Resolve(arg)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "rmdir: failed to remove '%s': %s\n", arg, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}

		if err := os.Remove(target); err != nil {
			fmt.Fprintf(ctx.Stderr, "rmdir: failed to remove '%s': %s\n", arg, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintf(ctx.Stdout, "rmdir: removing directory, '%s'\n", arg)
		}

		if parents {
			dir := filepath.Dir(target)
			for {
				if err := os.Remove(dir); err != nil {
					break
				}
				if verbose {
					fmt.Fprintf(ctx.Stdout, "rmdir: removing directory, '%s'\n", dir)
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}
	}
	return exit
}

func init() { command.Register(rmdirCommand{}) }
