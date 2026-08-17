package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type pwdCommand struct{}

func (pwdCommand) Name() string    { return "pwd" }
func (pwdCommand) Summary() string { return "print working directory" }

func (pwdCommand) Run(ctx *command.Context) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pwd: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	fmt.Fprintln(ctx.Stdout, wd)
	return command.ExitSuccess
}

func init() { command.Register(pwdCommand{}) }
