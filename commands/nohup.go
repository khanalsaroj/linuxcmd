package commands

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type nohupCommand struct{}

func (nohupCommand) Name() string    { return "nohup" }
func (nohupCommand) Summary() string { return "run a command immune to console hangups" }

func (nohupCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: nohup COMMAND [ARG...]")
		return command.ExitUsage
	}

	cmd := exec.Command(ctx.Args[0], ctx.Args[1:]...)
	cmd.Stdin = ctx.Stdin
	// CREATE_NEW_PROCESS_GROUP detaches the child from this console's
	// Ctrl+C group, the closest Windows equivalent of ignoring SIGHUP.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	if isInteractive(ctx) {
		f, err := os.OpenFile("nohup.out", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "nohup: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		defer f.Close()
		cmd.Stdout = f
		cmd.Stderr = f
		fmt.Fprintln(ctx.Stderr, "nohup: ignoring input and appending output to 'nohup.out'")
	} else {
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "nohup: %s\n", err)
		return command.ExitNotFound
	}
	return command.ExitSuccess
}

func init() { command.Register(nohupCommand{}) }
