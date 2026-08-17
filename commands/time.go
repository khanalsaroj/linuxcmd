package commands

import (
	"fmt"
	"os/exec"
	"time"

	"linuxcmd/internal/command"
)

type timeCommand struct{}

func (timeCommand) Name() string    { return "time" }
func (timeCommand) Summary() string { return "run a command and report elapsed time" }

func (timeCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: time COMMAND [ARG...]")
		return command.ExitUsage
	}

	cmd := exec.Command(ctx.Args[0], ctx.Args[1:]...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	fmt.Fprintf(ctx.Stderr, "\nreal\t%s\n", elapsed.Round(time.Millisecond))

	if runErr == nil {
		return command.ExitSuccess
	}
	if exitErr, ok := runErr.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	fmt.Fprintf(ctx.Stderr, "time: %s\n", runErr)
	return command.ExitNotFound
}

func init() { command.Register(timeCommand{}) }
