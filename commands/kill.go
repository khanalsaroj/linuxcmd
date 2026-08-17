package commands

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

// killCommand terminates a process by PID. Windows has no equivalent of
// Unix signals, so any "-SIGNAL"/"-9"/"-15" style flag is accepted and
// ignored (documented in the README): every kill maps onto
// TerminateProcess, which is unconditional, unlike a real SIGTERM that a
// process could choose to catch and handle gracefully.
type killCommand struct{}

func (killCommand) Name() string { return "kill" }
func (killCommand) Summary() string {
	return "terminate a process by PID (no signal distinction on Windows)"
}

func (killCommand) Run(ctx *command.Context) int {
	var pids []string
	for _, a := range ctx.Args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		pids = append(pids, a)
	}
	if len(pids) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: kill pid...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, p := range pids {
		pid, err := strconv.Atoi(p)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "kill: %s: arguments must be process IDs\n", p)
			exit = command.ExitFailure
			continue
		}

		handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "kill: (%d): %s\n", pid, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}
		if err := syscall.TerminateProcess(handle, 1); err != nil {
			fmt.Fprintf(ctx.Stderr, "kill: (%d): %s\n", pid, output.LinuxErrorText(err))
			exit = command.ExitFailure
		}
		syscall.CloseHandle(handle)
	}
	return exit
}

func init() { command.Register(killCommand{}) }
