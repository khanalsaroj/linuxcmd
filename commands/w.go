package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type wCommand struct{}

func (wCommand) Name() string    { return "w" }
func (wCommand) Summary() string { return "show who is logged in and system uptime" }

func (wCommand) Run(ctx *command.Context) int {
	d, err := systemUptime()
	if err == nil {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		fmt.Fprintf(ctx.Stdout, " up %d:%02d\n", hours, minutes)
	}

	names, err := enumerateSessionUsers()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "w: %s\n", err)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "%-16s %-10s %s\n", "USER", "TTY", "WHAT")
	for _, name := range names {
		fmt.Fprintf(ctx.Stdout, "%-16s %-10s %s\n", name, "console", "-")
	}
	return command.ExitSuccess
}

func init() { command.Register(wCommand{}) }
