package commands

import (
	"fmt"
	"os/exec"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// chgrpCommand has no true Windows equivalent: Windows ACLs have no
// POSIX-style single "primary group" concept. As the closest meaningful
// approximation, it grants the named group Modify access via icacls
// rather than reassigning any single "owning group" field.
type chgrpCommand struct{}

func (chgrpCommand) Name() string { return "chgrp" }
func (chgrpCommand) Summary() string {
	return "grant a group access (approximates chgrp; Windows has no primary-group field)"
}

func (chgrpCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: chgrp GROUP FILE...")
		return command.ExitUsage
	}
	group := ctx.Args[0]

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(ctx.Args[1:]) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "chgrp", arg, err)
			exit = command.ExitFailure
			continue
		}
		cmd := exec.Command("icacls.exe", resolved, "/grant", group+":M")
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(ctx.Stderr, "chgrp: %s: %s\n", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(chgrpCommand{}) }
