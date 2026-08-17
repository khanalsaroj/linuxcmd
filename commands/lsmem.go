package commands

import (
	"fmt"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
)

type lsmemCommand struct{}

func (lsmemCommand) Name() string    { return "lsmem" }
func (lsmemCommand) Summary() string { return "display memory ranges and totals" }

func (lsmemCommand) Run(ctx *command.Context) int {
	m, err := globalMemoryStatus()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "lsmem: %s\n", err)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stdout, "Total online memory:   %s\n", output.HumanSize(int64(m.TotalPhys)))
	fmt.Fprintf(ctx.Stdout, "Available memory:      %s\n", output.HumanSize(int64(m.AvailPhys)))
	return command.ExitSuccess
}

func init() { command.Register(lsmemCommand{}) }
