package commands

import (
	"fmt"
	"os/user"
	"strings"

	"linuxcmd/internal/command"
)

type groupsCommand struct{}

func (groupsCommand) Name() string    { return "groups" }
func (groupsCommand) Summary() string { return "print the current user's group memberships" }

func (groupsCommand) Run(ctx *command.Context) int {
	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "groups: %s\n", err)
		return command.ExitFailure
	}
	names := lookupGroupNames(u)
	if len(names) == 0 {
		fmt.Fprintln(ctx.Stdout, "users")
		return command.ExitSuccess
	}
	fmt.Fprintln(ctx.Stdout, strings.Join(names, " "))
	return command.ExitSuccess
}

func init() { command.Register(groupsCommand{}) }
