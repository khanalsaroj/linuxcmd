package commands

import (
	"fmt"
	"net"

	"linuxcmd/internal/command"
)

type hostCommand struct{}

func (hostCommand) Name() string    { return "host" }
func (hostCommand) Summary() string { return "look up a hostname's DNS records" }

func (hostCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: host NAME")
		return command.ExitUsage
	}
	name := ctx.Args[0]
	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Fprintf(ctx.Stdout, "host: %s not found\n", name)
		return command.ExitFailure
	}
	for _, a := range addrs {
		if net.ParseIP(a).To4() != nil {
			fmt.Fprintf(ctx.Stdout, "%s has address %s\n", name, a)
		} else {
			fmt.Fprintf(ctx.Stdout, "%s has IPv6 address %s\n", name, a)
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(hostCommand{}) }
