package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// dmesgCommand wraps wevtutil.exe to show recent System log entries, the
// closest native equivalent of a kernel-log view on Windows.
type dmesgCommand struct{}

func (dmesgCommand) Name() string    { return "dmesg" }
func (dmesgCommand) Summary() string { return "show recent system log messages" }

func (dmesgCommand) Run(ctx *command.Context) int {
	wevtutil := filepath.Join(systemRoot(), "System32", "wevtutil.exe")
	cmd := exec.Command(wevtutil, "qe", "System", "/c:50", "/rd:true", "/f:text")
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "dmesg: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(dmesgCommand{}) }
