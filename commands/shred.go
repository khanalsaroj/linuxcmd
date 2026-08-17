package commands

import (
	"crypto/rand"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// shredCommand overwrites a file with random data before optionally
// deleting it. This is best-effort only: on SSDs (wear leveling) and
// copy-on-write/journaled filesystems, the original data can still
// survive in blocks the overwrite never touches -- shred's own README
// warns about exactly this on Linux too, and it applies at least as much
// to NTFS on modern SSDs.
type shredCommand struct{}

func (shredCommand) Name() string    { return "shred" }
func (shredCommand) Summary() string { return "overwrite a file with random data before deleting" }

var shredSpec = parser.Spec{
	{Short: 'u', Long: "remove"},
	{Short: 'n', HasArg: true},
}

func (shredCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, shredSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "shred: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: shred [-u] [-n PASSES] FILE...")
		return command.ExitUsage
	}
	remove := res.Bool('u', "remove")
	passes := 3
	if v, ok := res.Value('n', ""); ok {
		n := 0
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "shred: invalid pass count '%s'\n", v)
			return command.ExitUsage
		}
		passes = n
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(res.Positional) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "shred", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "shred", arg, err)
			exit = command.ExitFailure
			continue
		}
		if err := shredFile(resolved, info.Size(), passes); err != nil {
			output.SimpleErrorf(ctx.Stderr, "shred", arg, err)
			exit = command.ExitFailure
			continue
		}
		if remove {
			if err := os.Remove(resolved); err != nil {
				output.SimpleErrorf(ctx.Stderr, "shred", arg, err)
				exit = command.ExitFailure
			}
		}
	}
	return exit
}

func shredFile(path string, size int64, passes int) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 64*1024)
	for p := 0; p < passes; p++ {
		if _, err := f.Seek(0, 0); err != nil {
			return err
		}
		remaining := size
		for remaining > 0 {
			n := int64(len(buf))
			if n > remaining {
				n = remaining
			}
			if _, err := rand.Read(buf[:n]); err != nil {
				return err
			}
			if _, err := f.Write(buf[:n]); err != nil {
				return err
			}
			remaining -= n
		}
	}
	return nil
}

func init() { command.Register(shredCommand{}) }
