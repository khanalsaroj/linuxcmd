package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// crontabCommand maps "-l" (list) onto schtasks.exe, the native Windows
// Task Scheduler CLI. Installing crontab-style recurring tasks from a
// crontab file isn't implemented; use "at" for one-time tasks instead.
type crontabCommand struct{}

func (crontabCommand) Name() string    { return "crontab" }
func (crontabCommand) Summary() string { return "list scheduled tasks (via Task Scheduler)" }

func schtasksExe() string {
	return filepath.Join(systemRoot(), "System32", "schtasks.exe")
}

func (crontabCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 || ctx.Args[0] != "-l" {
		fmt.Fprintln(ctx.Stderr, "usage: crontab -l")
		return command.ExitUsage
	}
	cmd := exec.Command(schtasksExe(), "/query", "/fo", "LIST")
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "crontab: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(crontabCommand{}) }
