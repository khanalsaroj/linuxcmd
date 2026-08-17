package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type dfCommand struct{}

func (dfCommand) Name() string    { return "df" }
func (dfCommand) Summary() string { return "report volume free/used space" }

var dfSpec = parser.Spec{
	{Short: 'h', Long: "human-readable"},
}

func (dfCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, dfSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "df: %s\n", err)
		return command.ExitUsage
	}
	human := res.Bool('h', "human-readable")

	volumes := res.Positional
	if len(volumes) == 0 {
		volumes = listDriveLetters()
	}

	format := func(n uint64) string {
		if human {
			return output.HumanSize(int64(n))
		}
		return fmt.Sprintf("%d", n/1024)
	}

	fmt.Fprintf(ctx.Stdout, "%-12s %10s %10s %10s %s\n", "Filesystem", "Size", "Used", "Avail", "Mounted")
	exit := command.ExitSuccess
	for _, v := range volumes {
		free, total, avail, err := diskFreeSpace(v)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "df", v, err)
			exit = command.ExitFailure
			continue
		}
		used := total - free
		fmt.Fprintf(ctx.Stdout, "%-12s %10s %10s %10s %s\n", v, format(total), format(used), format(avail), v)
	}
	return exit
}

func listDriveLetters() []string {
	var out []string
	for c := byte('A'); c <= 'Z'; c++ {
		root := string(c) + `:\`
		if _, err := os.Stat(root); err == nil {
			out = append(out, filepath.VolumeName(root)+`\`)
		}
	}
	if len(out) == 0 {
		out = []string{`C:\`}
	}
	return out
}

// diskFreeSpace wraps GetDiskFreeSpaceExW, returning (free-for-caller,
// total, available-to-caller) in bytes.
func diskFreeSpace(root string) (free, total, avail uint64, err error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDiskFreeSpaceExW")

	ptr, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return 0, 0, 0, err
	}

	var freeCaller, totalBytes, totalFree uint64
	ret, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeCaller)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if ret == 0 {
		return 0, 0, 0, callErr
	}
	return totalFree, totalBytes, freeCaller, nil
}

func init() { command.Register(dfCommand{}) }
