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

type installCommand struct{}

func (installCommand) Name() string    { return "install" }
func (installCommand) Summary() string { return "copy a file, creating parent directories" }

var installSpec = parser.Spec{
	{Short: 'D'},
	{Short: 'm', HasArg: true},
}

func (installCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, installSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "install: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: install [-D] [-m MODE] SOURCE DEST")
		return command.ExitUsage
	}
	createParents := res.Bool('D', "")
	src, dst := res.Positional[0], res.Positional[1]

	srcResolved, err := paths.Resolve(src)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "install", src, err)
		return command.ExitFailure
	}
	dstResolved, err := paths.Resolve(dst)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "install", dst, err)
		return command.ExitFailure
	}

	if createParents {
		if err := os.MkdirAll(filepath.Dir(dstResolved), 0755); err != nil {
			output.SimpleErrorf(ctx.Stderr, "install", dst, err)
			return command.ExitFailure
		}
	}

	if err := fsutil.CopyFile(srcResolved, dstResolved); err != nil {
		output.SimpleErrorf(ctx.Stderr, "install", src, err)
		return command.ExitFailure
	}

	if v, ok := res.Value('m', ""); ok {
		writable, hasEffect, err := chmodWritable(v, true)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "install: %s\n", err)
			return command.ExitFailure
		}
		if hasEffect {
			perm := os.FileMode(0444)
			if writable {
				perm = 0644
			}
			os.Chmod(dstResolved, perm)
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(installCommand{}) }
