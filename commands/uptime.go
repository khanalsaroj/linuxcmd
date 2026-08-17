package commands

import (
	"fmt"
	"syscall"
	"time"

	"linuxcmd/internal/command"
)

type uptimeCommand struct{}

func (uptimeCommand) Name() string    { return "uptime" }
func (uptimeCommand) Summary() string { return "report how long the system has been running" }

func systemUptime() (time.Duration, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetTickCount64")
	ret, _, _ := proc.Call()
	return time.Duration(ret) * time.Millisecond, nil
}

func (uptimeCommand) Run(ctx *command.Context) int {
	d, err := systemUptime()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "uptime: %s\n", err)
		return command.ExitFailure
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	fmt.Fprintf(ctx.Stdout, "up %d:%02d, load average: unavailable\n", hours, minutes)
	return command.ExitSuccess
}

func init() { command.Register(uptimeCommand{}) }
