package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type hostnameCommand struct{}

func (hostnameCommand) Name() string    { return "hostname" }
func (hostnameCommand) Summary() string { return "print the system hostname" }

func (hostnameCommand) Run(ctx *command.Context) int {
	name, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "hostname: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	fmt.Fprintln(ctx.Stdout, name)
	return command.ExitSuccess
}

func init() { command.Register(hostnameCommand{}) }
