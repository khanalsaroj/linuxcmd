package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// netstatCommand wraps Windows' own netstat.exe, whose flags (-a -n -o)
// already read naturally to a Linux user, rather than reimplementing
// socket-table enumeration via the IP Helper API.
type netstatCommand struct{}

func (netstatCommand) Name() string    { return "netstat" }
func (netstatCommand) Summary() string { return "show network connections and listening sockets" }

func (netstatCommand) Run(ctx *command.Context) int {
	winArgs := ctx.Args
	if len(winArgs) == 0 {
		winArgs = []string{"-a", "-n"}
	}
	netstatExe := filepath.Join(systemRoot(), "System32", "netstat.exe")
	cmd := exec.Command(netstatExe, winArgs...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "netstat: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(netstatCommand{}) }
