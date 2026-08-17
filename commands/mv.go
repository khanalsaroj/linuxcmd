package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"linuxcmd/internal/command"
	"linuxcmd/internal/fsutil"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type mvCommand struct{}

func (mvCommand) Name() string    { return "mv" }
func (mvCommand) Summary() string { return "move/rename files and directories" }

var mvSpec = parser.Spec{
	{Short: 'f', Long: "force"},
	{Short: 'n', Long: "no-clobber"},
	{Short: 'v', Long: "verbose"},
}

func (mvCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, mvSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mv: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) < 2 {
		fmt.Fprintln(ctx.Stderr, "mv: missing file operand")
		return command.ExitUsage
	}

	noClobber := res.Bool('n', "no-clobber")
	verbose := res.Bool('v', "verbose")

	sources := paths.ExpandGlobs(res.Positional[:len(res.Positional)-1])
	dest := res.Positional[len(res.Positional)-1]

	destResolved, err := paths.Resolve(dest)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mv: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	destInfo, statErr := os.Stat(destResolved)
	destIsDir := statErr == nil && destInfo.IsDir()

	if len(sources) > 1 && !destIsDir {
		fmt.Fprintf(ctx.Stderr, "mv: target '%s' is not a directory\n", dest)
		return command.ExitFailure
	}

	exit := command.ExitSuccess
	for _, src := range sources {
		srcResolved, err := paths.Resolve(src)
		if err != nil {
			output.Errorf(ctx.Stderr, "mv", "cannot stat", src, err)
			exit = command.ExitFailure
			continue
		}

		finalDest := destResolved
		if destIsDir {
			finalDest = filepath.Join(destResolved, filepath.Base(srcResolved))
		}

		if noClobber {
			if _, err := os.Stat(finalDest); err == nil {
				continue
			}
		}

		if err := fsutil.Move(srcResolved, finalDest); err != nil {
			output.Errorf(ctx.Stderr, "mv", "cannot move", src, err)
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintf(ctx.Stdout, "'%s' -> '%s'\n", src, finalDest)
		}
	}
	return exit
}

func init() { command.Register(mvCommand{}) }
