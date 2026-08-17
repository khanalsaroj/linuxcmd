package commands

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
)

type sttyCommand struct{}

func (sttyCommand) Name() string    { return "stty" }
func (sttyCommand) Summary() string { return "report or change console mode flags" }

func getConsoleMode(h syscall.Handle) (uint32, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleMode")
	var mode uint32
	ret, _, err := proc.Call(uintptr(h), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return 0, err
	}
	return mode, nil
}

func (sttyCommand) Run(ctx *command.Context) int {
	f, ok := ctx.Stdin.(*os.File)
	if !ok {
		fmt.Fprintln(ctx.Stderr, "stty: standard input is not a console")
		return command.ExitFailure
	}
	mode, err := getConsoleMode(syscall.Handle(f.Fd()))
	if err != nil {
		fmt.Fprintln(ctx.Stderr, "stty: standard input is not a console")
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "console mode: 0x%08x\n", mode)
	return command.ExitSuccess
}

func init() { command.Register(sttyCommand{}) }
