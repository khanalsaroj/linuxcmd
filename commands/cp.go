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

type cpCommand struct{}

func (cpCommand) Name() string    { return "cp" }
func (cpCommand) Summary() string { return "copy files and directories" }

var cpSpec = parser.Spec{
	{Short: 'r'},
	{Short: 'R'},
	{Short: 'f', Long: "force"},
	{Short: 'n', Long: "no-clobber"},
	{Short: 'v', Long: "verbose"},
}

func (cpCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, cpSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cp: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) < 2 {
		fmt.Fprintln(ctx.Stderr, "cp: missing file operand")
		return command.ExitUsage
	}

	recursive := res.Bool('r', "") || res.Bool('R', "")
	noClobber := res.Bool('n', "no-clobber")
	verbose := res.Bool('v', "verbose")

	sources := paths.ExpandGlobs(res.Positional[:len(res.Positional)-1])
	dest := res.Positional[len(res.Positional)-1]

	destResolved, err := paths.Resolve(dest)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cp: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	destInfo, statErr := os.Stat(destResolved)
	destIsDir := statErr == nil && destInfo.IsDir()

	if len(sources) > 1 && !destIsDir {
		fmt.Fprintf(ctx.Stderr, "cp: target '%s' is not a directory\n", dest)
		return command.ExitFailure
	}

	exit := command.ExitSuccess
	for _, src := range sources {
		srcResolved, err := paths.Resolve(src)
		if err != nil {
			output.Errorf(ctx.Stderr, "cp", "cannot stat", src, err)
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

		srcInfo, err := os.Stat(srcResolved)
		if err != nil {
			output.Errorf(ctx.Stderr, "cp", "cannot stat", src, err)
			exit = command.ExitFailure
			continue
		}

		if srcInfo.IsDir() {
			if !recursive {
				fmt.Fprintf(ctx.Stderr, "cp: -r not specified; omitting directory '%s'\n", src)
				exit = command.ExitFailure
				continue
			}
			err = fsutil.CopyRecursive(srcResolved, finalDest)
		} else {
			err = fsutil.CopyFile(srcResolved, finalDest)
		}
		if err != nil {
			output.Errorf(ctx.Stderr, "cp", "cannot copy", src, err)
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintf(ctx.Stdout, "'%s' -> '%s'\n", src, finalDest)
		}
	}
	return exit
}

func init() { command.Register(cpCommand{}) }
