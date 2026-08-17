package commands

import (
	"fmt"
	"net"
	"os/user"

	"linuxcmd/internal/command"
)

type getentCommand struct{}

func (getentCommand) Name() string    { return "getent" }
func (getentCommand) Summary() string { return "query hosts, passwd, or group databases" }

func (getentCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 1 {
		fmt.Fprintln(ctx.Stderr, "usage: getent DATABASE [KEY]")
		return command.ExitUsage
	}
	database := ctx.Args[0]

	switch database {
	case "hosts":
		if len(ctx.Args) < 2 {
			fmt.Fprintln(ctx.Stderr, "usage: getent hosts NAME")
			return command.ExitUsage
		}
		addrs, err := net.LookupHost(ctx.Args[1])
		if err != nil {
			return command.ExitFailure
		}
		for _, a := range addrs {
			fmt.Fprintf(ctx.Stdout, "%-16s %s\n", a, ctx.Args[1])
		}
		return command.ExitSuccess

	case "passwd":
		u, err := user.Current()
		if err != nil {
			return command.ExitFailure
		}
		if len(ctx.Args) >= 2 && ctx.Args[1] != u.Username {
			return command.ExitFailure
		}
		fmt.Fprintf(ctx.Stdout, "%s:x:%s:%s::%s:\n", u.Username, u.Uid, u.Gid, u.HomeDir)
		return command.ExitSuccess

	case "group":
		u, err := user.Current()
		if err != nil {
			return command.ExitFailure
		}
		names := lookupGroupNames(u)
		name := primaryGroupName(names)
		if len(ctx.Args) >= 2 && ctx.Args[1] != name {
			return command.ExitFailure
		}
		fmt.Fprintf(ctx.Stdout, "%s:x:%s:\n", name, u.Gid)
		return command.ExitSuccess

	default:
		fmt.Fprintf(ctx.Stderr, "getent: unknown database '%s'\n", database)
		return command.ExitUsage
	}
}

func init() { command.Register(getentCommand{}) }
