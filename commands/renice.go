package commands

import (
	"fmt"
	"strconv"
	"syscall"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type reniceCommand struct{}

func (reniceCommand) Name() string    { return "renice" }
func (reniceCommand) Summary() string { return "change a running process's scheduling priority" }

var reniceSpec = parser.Spec{
	{Short: 'p', HasArg: true},
}

var setPriorityClassProc = syscall.NewLazyDLL("kernel32.dll").NewProc("SetPriorityClass")

func (reniceCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, reniceSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "renice: %s\n", err)
		return command.ExitUsage
	}
	pidStr, ok := res.Value('p', "")
	if !ok || len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: renice PRIORITY -p PID")
		return command.ExitUsage
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "renice: invalid PID '%s'\n", pidStr)
		return command.ExitUsage
	}
	niceVal, err := strconv.Atoi(res.Positional[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "renice: invalid priority '%s'\n", res.Positional[0])
		return command.ExitUsage
	}

	const processSetInformation = 0x0200
	handle, err := syscall.OpenProcess(processSetInformation, false, uint32(pid))
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "renice: (%d): %s\n", pid, output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer syscall.CloseHandle(handle)

	ret, _, callErr := setPriorityClassProc.Call(uintptr(handle), uintptr(niceToPriorityClass(niceVal)))
	if ret == 0 {
		fmt.Fprintf(ctx.Stderr, "renice: (%d): %s\n", pid, callErr)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "%d: old priority unavailable, new priority class set\n", pid)
	return command.ExitSuccess
}

func init() { command.Register(reniceCommand{}) }
