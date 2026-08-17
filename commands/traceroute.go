package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// tracerouteCommand wraps Windows' own tracert.exe rather than
// reimplementing incrementing-TTL ICMP probes: like ping, raw ICMP needs
// elevated privileges on Windows, and tracert.exe already does exactly
// this.
type tracerouteCommand struct{}

func (tracerouteCommand) Name() string    { return "traceroute" }
func (tracerouteCommand) Summary() string { return "trace the network path to a host" }

func (tracerouteCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: traceroute HOST")
		return command.ExitUsage
	}
	host := ctx.Args[len(ctx.Args)-1]

	tracertExe := filepath.Join(systemRoot(), "System32", "tracert.exe")
	cmd := exec.Command(tracertExe, host)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "traceroute: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(tracerouteCommand{}) }
