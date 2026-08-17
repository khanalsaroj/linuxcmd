package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
)

type ttyCommand struct{}

func (ttyCommand) Name() string    { return "tty" }
func (ttyCommand) Summary() string { return "print the terminal connected to standard input" }

func (ttyCommand) Run(ctx *command.Context) int {
	f, ok := ctx.Stdin.(*os.File)
	if !ok {
		fmt.Fprintln(ctx.Stdout, "not a tty")
		return command.ExitFailure
	}
	stat, err := f.Stat()
	if err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		fmt.Fprintln(ctx.Stdout, "not a tty")
		return command.ExitFailure
	}
	fmt.Fprintln(ctx.Stdout, "CONIN$")
	return command.ExitSuccess
}

func init() { command.Register(ttyCommand{}) }
