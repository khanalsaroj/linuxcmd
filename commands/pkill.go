package commands

import (
	"fmt"
	"strings"
	"syscall"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type pkillCommand struct{}

func (pkillCommand) Name() string    { return "pkill" }
func (pkillCommand) Summary() string { return "terminate processes by name" }

var pkillSpec = parser.Spec{
	{Short: 'f'},
}

func (pkillCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, pkillSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pkill: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: pkill PATTERN")
		return command.ExitUsage
	}
	pattern := strings.ToLower(res.Positional[0])

	procs, err := snapshotProcesses()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pkill: %s\n", err)
		return command.ExitFailure
	}

	matched := false
	exit := command.ExitSuccess
	for _, p := range procs {
		if !strings.Contains(strings.ToLower(p.Name), pattern) {
			continue
		}
		matched = true
		handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, p.PID)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "pkill: (%d): %s\n", p.PID, output.LinuxErrorText(err))
			exit = command.ExitFailure
			continue
		}
		if err := syscall.TerminateProcess(handle, 1); err != nil {
			fmt.Fprintf(ctx.Stderr, "pkill: (%d): %s\n", p.PID, output.LinuxErrorText(err))
			exit = command.ExitFailure
		} else {
			fmt.Fprintf(ctx.Stdout, "%s (%d) killed\n", p.Name, p.PID)
		}
		syscall.CloseHandle(handle)
	}
	if !matched {
		return command.ExitFailure
	}
	return exit
}

func init() { command.Register(pkillCommand{}) }
