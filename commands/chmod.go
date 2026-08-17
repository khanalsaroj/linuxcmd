package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// chmodCommand maps chmod's mode argument onto Windows' single
// read-only file attribute, the only piece of POSIX permission bits
// Windows actually tracks (see output.FormatMode). "+x"/"-x" are
// accepted but have no effect, since Windows has no separate execute
// bit; this is documented in the README rather than silently faked.
type chmodCommand struct{}

func (chmodCommand) Name() string { return "chmod" }
func (chmodCommand) Summary() string {
	return "change file write permission (Windows read-only attribute)"
}

// chmodWritable reports whether mode grants write access, or ok=false if
// mode has no effect on writability (bare "+x"/"-x").
func chmodWritable(mode string, currentWritable bool) (writable bool, ok bool, err error) {
	switch mode {
	case "+w":
		return true, true, nil
	case "-w":
		return false, true, nil
	case "+x", "-x":
		return currentWritable, false, nil
	}
	n, parseErr := strconv.ParseUint(mode, 8, 32)
	if parseErr != nil {
		return false, false, fmt.Errorf("invalid mode: '%s'", mode)
	}
	return n&0200 != 0, true, nil
}

func (chmodCommand) Run(ctx *command.Context) int {
	// MODE can itself look like a flag (e.g. "-w"), so this scans only
	// for a leading -R/--recursive rather than using the shared flag
	// parser, the same way xargs avoids it for its wrapped command.
	args := ctx.Args
	recursive := false
	i := 0
	for i < len(args) && (args[i] == "-R" || args[i] == "--recursive") {
		recursive = true
		i++
	}
	if len(args)-i < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: chmod [-R] MODE FILE...")
		return command.ExitUsage
	}
	mode := args[i]
	positional := args[i+1:]

	apply := func(path string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		writable, ok, err := chmodWritable(mode, info.Mode()&0200 != 0)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		perm := os.FileMode(0444)
		if writable {
			perm = 0644
		}
		return os.Chmod(path, perm)
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(positional) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.Errorf(ctx.Stderr, "chmod", "cannot access", arg, err)
			exit = command.ExitFailure
			continue
		}
		if !recursive {
			if err := apply(resolved); err != nil {
				output.Errorf(ctx.Stderr, "chmod", "changing permissions of", arg, err)
				exit = command.ExitFailure
			}
			continue
		}
		walkErr := filepath.Walk(resolved, func(p string, _ os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			return apply(p)
		})
		if walkErr != nil {
			output.Errorf(ctx.Stderr, "chmod", "changing permissions of", arg, walkErr)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(chmodCommand{}) }
