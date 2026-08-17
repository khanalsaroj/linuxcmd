package commands

import (
	"fmt"
	"os"
	"strconv"

	"linuxcmd/internal/command"
)

// umaskCommand reports or validates a umask-style octal mask.
// Unlike a real shell builtin, this process can't change the umask
// inherited by the parent CMD session or later linuxcmd invocations, the
// same limitation "cd" documents in the README: each command runs as its
// own short-lived process. Setting a value here only prints confirmation
// and validates the syntax.
type umaskCommand struct{}

func (umaskCommand) Name() string { return "umask" }
func (umaskCommand) Summary() string {
	return "report a compatibility file-creation mask (does not persist to the parent shell)"
}

const defaultUmask = "0022"

func (umaskCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		mask := os.Getenv("LINUXCMD_UMASK")
		if mask == "" {
			mask = defaultUmask
		}
		fmt.Fprintln(ctx.Stdout, mask)
		return command.ExitSuccess
	}

	arg := ctx.Args[0]
	if _, err := strconv.ParseUint(arg, 8, 32); err != nil {
		fmt.Fprintf(ctx.Stderr, "umask: '%s' is not a valid octal mask\n", arg)
		return command.ExitUsage
	}
	fmt.Fprintf(ctx.Stderr, "umask: set to %s for this process only; it cannot persist to the parent shell\n", arg)
	return command.ExitSuccess
}

func init() { command.Register(umaskCommand{}) }
