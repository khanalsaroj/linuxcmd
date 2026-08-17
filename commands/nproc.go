package commands

import (
	"fmt"
	"runtime"

	"linuxcmd/internal/command"
)

type nprocCommand struct{}

func (nprocCommand) Name() string    { return "nproc" }
func (nprocCommand) Summary() string { return "print the number of available processors" }

func (nprocCommand) Run(ctx *command.Context) int {
	fmt.Fprintln(ctx.Stdout, runtime.NumCPU())
	return command.ExitSuccess
}

func init() { command.Register(nprocCommand{}) }
