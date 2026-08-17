package commands

import (
	"fmt"
	"strings"

	"linuxcmd/internal/command"
)

type usersCommand struct{}

func (usersCommand) Name() string    { return "users" }
func (usersCommand) Summary() string { return "print interactive session usernames" }

func (usersCommand) Run(ctx *command.Context) int {
	names, err := enumerateSessionUsers()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "users: %s\n", err)
		return command.ExitFailure
	}
	fmt.Fprintln(ctx.Stdout, strings.Join(names, " "))
	return command.ExitSuccess
}

func init() { command.Register(usersCommand{}) }
