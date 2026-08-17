package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// cdCommand resolves and validates a target directory, printing the
// resolved absolute Windows path on success.
//
// It deliberately does NOT (and structurally cannot) change the calling
// shell's working directory: a child process can never alter its
// parent's current directory on any operating system, since each process
// only has its own copy of that state. This is exactly why every real
// Unix shell implements "cd" as a builtin rather than an external
// program, and why cmd.exe does the same. This binary exists so it can
// be composed by the optional DOSKEY macro layer (installed via
// `install.ps1 -EnableDoskeyOverrides`), which runs cd's *actual*
// directory change through cmd.exe's own builtin in the same process,
// after asking this binary to resolve Linux-style syntax (~, /tmp, ...)
// into a real Windows path first. See README.md's "cd limitation"
// section.
type cdCommand struct{}

func (cdCommand) Name() string { return "cd" }
func (cdCommand) Summary() string {
	return "resolve a directory path (see README: cannot change the parent shell's directory on its own)"
}

func (cdCommand) Run(ctx *command.Context) int {
	target := "~"
	if len(ctx.Args) > 0 {
		target = ctx.Args[0]
	}

	resolved, err := paths.Resolve(target)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cd: %s: %s\n", target, output.LinuxErrorText(err))
		return command.ExitFailure
	}

	info, err := os.Stat(resolved)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cd: %s: %s\n", target, output.LinuxErrorText(err))
		return command.ExitFailure
	}
	if !info.IsDir() {
		fmt.Fprintf(ctx.Stderr, "cd: %s: Not a directory\n", target)
		return command.ExitFailure
	}

	fmt.Fprintln(ctx.Stdout, resolved)
	return command.ExitSuccess
}

func init() { command.Register(cdCommand{}) }
