package commands

import (
	"fmt"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type freeCommand struct{}

func (freeCommand) Name() string    { return "free" }
func (freeCommand) Summary() string { return "report physical and virtual memory usage" }

var freeSpec = parser.Spec{
	{Short: 'h', Long: "human"},
}

// memoryStatusEx mirrors the Win32 MEMORYSTATUSEX struct expected by
// GlobalMemoryStatusEx.
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func globalMemoryStatus() (memoryStatusEx, error) {
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GlobalMemoryStatusEx")
	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&m)))
	if ret == 0 {
		return m, err
	}
	return m, nil
}

func (freeCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, freeSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "free: %s\n", err)
		return command.ExitUsage
	}
	human := res.Bool('h', "human")

	m, err := globalMemoryStatus()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "free: %s\n", err)
		return command.ExitFailure
	}

	format := func(n uint64) string {
		if human {
			return output.HumanSize(int64(n))
		}
		return fmt.Sprintf("%d", n/1024)
	}

	used := m.TotalPhys - m.AvailPhys
	pageUsed := m.TotalPageFile - m.AvailPageFile

	fmt.Fprintf(ctx.Stdout, "%-8s %10s %10s %10s\n", "", "total", "used", "free")
	fmt.Fprintf(ctx.Stdout, "%-8s %10s %10s %10s\n", "Mem:", format(m.TotalPhys), format(used), format(m.AvailPhys))
	fmt.Fprintf(ctx.Stdout, "%-8s %10s %10s %10s\n", "Swap:", format(m.TotalPageFile), format(pageUsed), format(m.AvailPageFile))
	return command.ExitSuccess
}

func init() { command.Register(freeCommand{}) }
