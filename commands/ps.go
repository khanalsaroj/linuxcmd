package commands

import (
	"fmt"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
)

type psCommand struct{}

func (psCommand) Name() string    { return "ps" }
func (psCommand) Summary() string { return "report a snapshot of current processes" }

const th32csSnapProcess = 0x00000002

func (psCommand) Run(ctx *command.Context) int {
	snapshot, err := syscall.CreateToolhelp32Snapshot(th32csSnapProcess, 0)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ps: %s\n", err)
		return command.ExitFailure
	}
	defer syscall.CloseHandle(snapshot)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	fmt.Fprintf(ctx.Stdout, "%8s %8s %s\n", "PID", "PPID", "CMD")
	for err = syscall.Process32First(snapshot, &entry); err == nil; err = syscall.Process32Next(snapshot, &entry) {
		name := syscall.UTF16ToString(entry.ExeFile[:])
		fmt.Fprintf(ctx.Stdout, "%8d %8d %s\n", entry.ProcessID, entry.ParentProcessID, name)
	}
	return command.ExitSuccess
}

func init() { command.Register(psCommand{}) }
