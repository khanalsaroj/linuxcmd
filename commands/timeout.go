package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"linuxcmd/internal/command"
)

// killProcessTree terminates pid and its descendants. cmd.Process.Kill
// only terminates the direct child; when that child is cmd.exe or a
// shell, its own children (e.g. ping.exe launched via "cmd /c ping ...")
// survive as orphans and keep any inherited stdout/stderr pipes open,
// which would otherwise make cmd.Wait() block until they exit on their
// own. taskkill's /T recurses through the whole tree.
func killProcessTree(pid int) {
	exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run()
}

type timeoutCommand struct{}

func (timeoutCommand) Name() string    { return "timeout" }
func (timeoutCommand) Summary() string { return "run a command with a time limit" }

func (timeoutCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: timeout DURATION COMMAND [ARG...]")
		return command.ExitUsage
	}
	d, err := parseSleepDuration(ctx.Args[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "timeout: invalid time interval '%s'\n", ctx.Args[0])
		return command.ExitUsage
	}

	cmd := exec.Command(ctx.Args[1], ctx.Args[2:]...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(ctx.Stderr, "timeout: %s\n", err)
		return command.ExitNotFound
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err == nil {
			return command.ExitSuccess
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return command.ExitFailure
	case <-time.After(d):
		killProcessTree(cmd.Process.Pid)
		<-done
		return 124 // matches GNU timeout's exit status when it kills the child
	}
}

func init() { command.Register(timeoutCommand{}) }
