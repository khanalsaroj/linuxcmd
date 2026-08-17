package commands

import (
	"fmt"
	"os/exec"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// chownCommand wraps icacls.exe /setowner, since taking real ownership
// of a file requires the Windows security-descriptor APIs (SetNamedSecurityInfo)
// that icacls already exposes as a stable CLI, and typically needs
// administrator or SeTakeOwnershipPrivilege rights -- icacls itself
// enforces that, so its error is surfaced as-is.
type chownCommand struct{}

func (chownCommand) Name() string    { return "chown" }
func (chownCommand) Summary() string { return "change file owner (wraps icacls /setowner)" }

func (chownCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: chown USER FILE...")
		return command.ExitUsage
	}
	owner := ctx.Args[0]

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(ctx.Args[1:]) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "chown", arg, err)
			exit = command.ExitFailure
			continue
		}
		cmd := exec.Command("icacls.exe", resolved, "/setowner", owner)
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(ctx.Stderr, "chown: %s: %s\n", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(chownCommand{}) }
