package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// arpCommand wraps Windows' own arp.exe; "arp -a" already means the same
// thing on both platforms.
type arpCommand struct{}

func (arpCommand) Name() string    { return "arp" }
func (arpCommand) Summary() string { return "show the ARP/neighbor table" }

func (arpCommand) Run(ctx *command.Context) int {
	winArgs := ctx.Args
	if len(winArgs) == 0 {
		winArgs = []string{"-a"}
	}
	arpExe := filepath.Join(systemRoot(), "System32", "arp.exe")
	cmd := exec.Command(arpExe, winArgs...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "arp: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(arpCommand{}) }
