package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type archCommand struct{}

func (archCommand) Name() string    { return "arch" }
func (archCommand) Summary() string { return "print machine architecture" }

func (archCommand) Run(ctx *command.Context) int {
	fmt.Fprintln(ctx.Stdout, unameArch())
	return command.ExitSuccess
}

func init() { command.Register(archCommand{}) }
