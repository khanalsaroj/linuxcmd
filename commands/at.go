package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"linuxcmd/internal/command"
)

// atCommand creates a one-time Task Scheduler task via schtasks.exe,
// Windows' closest equivalent of Unix "at".
type atCommand struct{}

func (atCommand) Name() string    { return "at" }
func (atCommand) Summary() string { return "schedule a one-time command via Task Scheduler" }

func (atCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: at HH:MM COMMAND [ARG...]")
		return command.ExitUsage
	}
	when := ctx.Args[0]
	command_ := strings.Join(ctx.Args[1:], " ")
	taskName := "linuxcmd-at-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	cmd := exec.Command(schtasksExe(), "/create", "/tn", taskName, "/tr", command_, "/sc", "once", "/st", when, "/f")
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "at: %s\n", err)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "job scheduled as task '%s'\n", taskName)
	return command.ExitSuccess
}

func init() { command.Register(atCommand{}) }
