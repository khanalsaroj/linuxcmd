package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
)

// hostnamectlCommand supports only "status" (the common read path).
// Renaming a Windows computer needs a reboot and administrator rights via
// a different API entirely (SetComputerNameEx + restart), so
// "set-hostname" is reported as unsupported rather than half-implemented.
type hostnamectlCommand struct{}

func (hostnamectlCommand) Name() string    { return "hostnamectl" }
func (hostnamectlCommand) Summary() string { return "show host metadata" }

func (hostnamectlCommand) Run(ctx *command.Context) int {
	sub := "status"
	if len(ctx.Args) > 0 {
		sub = ctx.Args[0]
	}
	if sub != "status" {
		fmt.Fprintf(ctx.Stderr, "hostnamectl: '%s' is not supported; only 'status' is implemented\n", sub)
		return command.ExitUsage
	}

	host, _ := os.Hostname()
	fmt.Fprintf(ctx.Stdout, "   Static hostname: %s\n", host)
	fmt.Fprintf(ctx.Stdout, "  Operating System: %s\n", "Windows_NT")
	fmt.Fprintf(ctx.Stdout, "       Architecture: %s\n", unameArch())
	return command.ExitSuccess
}

func init() { command.Register(hostnamectlCommand{}) }
