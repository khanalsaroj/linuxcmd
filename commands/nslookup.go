package commands

import (
	"fmt"
	"net"

	"linuxcmd/internal/command"
)

type nslookupCommand struct{}

func (nslookupCommand) Name() string    { return "nslookup" }
func (nslookupCommand) Summary() string { return "query DNS for a name's addresses" }

func (nslookupCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: nslookup NAME")
		return command.ExitUsage
	}
	name := ctx.Args[0]
	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "nslookup: can't resolve '%s'\n", name)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "Name:\t%s\n", name)
	for _, a := range addrs {
		fmt.Fprintf(ctx.Stdout, "Address: %s\n", a)
	}
	return command.ExitSuccess
}

func init() { command.Register(nslookupCommand{}) }
