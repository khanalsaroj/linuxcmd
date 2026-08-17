package commands

import (
	"fmt"
	"net"
	"strings"

	"linuxcmd/internal/command"
)

// ipCommand implements a read-only subset of iproute2's "ip" — "addr"
// (the default) and "link" — using Go's standard net package instead of
// shelling out, since Windows has no native "ip" tool to wrap the way
// ping.exe exists for "ping". Route/address *modification* subcommands
// ("ip addr add", "ip route") are out of scope for this MVP.
type ipCommand struct{}

func (ipCommand) Name() string    { return "ip" }
func (ipCommand) Summary() string { return "show network interfaces and addresses" }

func (ipCommand) Run(ctx *command.Context) int {
	sub := "addr"
	if len(ctx.Args) > 0 {
		sub = ctx.Args[0]
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ip: %s\n", err)
		return command.ExitFailure
	}

	switch sub {
	case "addr", "a", "address":
		for _, iface := range ifaces {
			printLink(ctx, iface)
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				family := "inet"
				if strings.Contains(a.String(), ":") {
					family = "inet6"
				}
				fmt.Fprintf(ctx.Stdout, "    %s %s\n", family, a.String())
			}
		}
	case "link", "l":
		for _, iface := range ifaces {
			printLink(ctx, iface)
		}
	default:
		fmt.Fprintf(ctx.Stderr, "ip: unknown object \"%s\" (supported: addr, link)\n", sub)
		return command.ExitUsage
	}
	return command.ExitSuccess
}

func printLink(ctx *command.Context, iface net.Interface) {
	flags := strings.ToUpper(strings.ReplaceAll(iface.Flags.String(), "|", ","))
	fmt.Fprintf(ctx.Stdout, "%d: %s: <%s> mtu %d\n", iface.Index, iface.Name, flags, iface.MTU)
	if iface.HardwareAddr != nil && len(iface.HardwareAddr) > 0 {
		fmt.Fprintf(ctx.Stdout, "    link/ether %s\n", iface.HardwareAddr.String())
	}
}

func init() { command.Register(ipCommand{}) }
