package commands

import (
	"fmt"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type whoamiCommand struct{}

func (whoamiCommand) Name() string    { return "whoami" }
func (whoamiCommand) Summary() string { return "print the current user name" }

func (whoamiCommand) Run(ctx *command.Context) int {
	fmt.Fprintln(ctx.Stdout, output.CurrentUsername())
	return command.ExitSuccess
}

func init() { command.Register(whoamiCommand{}) }
