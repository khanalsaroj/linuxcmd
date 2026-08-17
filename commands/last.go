package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// lastCommand wraps Windows' own wevtutil.exe to read sign-in (logon,
// event ID 4624) records from the Security event log, the closest native
// equivalent of Linux "last". Reading the Security log typically
// requires administrator rights, so a permission error here is expected
// on a standard user account and is surfaced as-is rather than hidden.
type lastCommand struct{}

func (lastCommand) Name() string    { return "last" }
func (lastCommand) Summary() string { return "show recent sign-in events from the Security event log" }

func (lastCommand) Run(ctx *command.Context) int {
	wevtutil := filepath.Join(systemRoot(), "System32", "wevtutil.exe")
	args := []string{
		"qe", "Security",
		"/q:*[System[(EventID=4624)]]",
		"/c:10", "/rd:true", "/f:text",
	}
	cmd := exec.Command(wevtutil, args...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "last: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(lastCommand{}) }
