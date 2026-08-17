package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type whoCommand struct{}

func (whoCommand) Name() string    { return "who" }
func (whoCommand) Summary() string { return "list active sessions" }

func (whoCommand) Run(ctx *command.Context) int {
	names, err := enumerateSessionUsers()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "who: %s\n", err)
		return command.ExitFailure
	}
	for _, name := range names {
		fmt.Fprintf(ctx.Stdout, "%-16s console\n", name)
	}
	return command.ExitSuccess
}

func init() { command.Register(whoCommand{}) }
