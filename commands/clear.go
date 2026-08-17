package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type clearCommand struct{}

func (clearCommand) Name() string    { return "clear" }
func (clearCommand) Summary() string { return "clear the terminal screen" }

func (clearCommand) Run(ctx *command.Context) int {
	enableVirtualTerminal()
	// ANSI: clear visible screen + scrollback, then home the cursor.
	fmt.Fprint(ctx.Stdout, "\x1b[2J\x1b[3J\x1b[H")
	return command.ExitSuccess
}

func init() { command.Register(clearCommand{}) }
