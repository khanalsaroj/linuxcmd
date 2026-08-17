package commands

import (
	"fmt"
	"os/user"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type idCommand struct{}

func (idCommand) Name() string    { return "id" }
func (idCommand) Summary() string { return "print user and group identity" }

func (idCommand) Run(ctx *command.Context) int {
	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "id: %s\n", err)
		return command.ExitFailure
	}
	name := output.CurrentUsername()

	groupNames := lookupGroupNames(u)
	fmt.Fprintf(ctx.Stdout, "uid=%s(%s) gid=%s(%s)", u.Uid, name, u.Gid, primaryGroupName(groupNames))
	if len(groupNames) > 0 {
		fmt.Fprint(ctx.Stdout, " groups=")
		for i, g := range groupNames {
			if i > 0 {
				fmt.Fprint(ctx.Stdout, ",")
			}
			fmt.Fprint(ctx.Stdout, g)
		}
	}
	fmt.Fprintln(ctx.Stdout)
	return command.ExitSuccess
}

func lookupGroupNames(u *user.User) []string {
	gids, err := u.GroupIds()
	if err != nil {
		return nil
	}
	var names []string
	for _, gid := range gids {
		if g, err := user.LookupGroupId(gid); err == nil {
			names = append(names, g.Name)
		}
	}
	return names
}

func primaryGroupName(groups []string) string {
	if len(groups) == 0 {
		return "users"
	}
	return groups[0]
}

func init() { command.Register(idCommand{}) }
