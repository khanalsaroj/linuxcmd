package commands

import (
	"fmt"
	"runtime"

	"linuxcmd/internal/command"
)

type getconfCommand struct{}

func (getconfCommand) Name() string    { return "getconf" }
func (getconfCommand) Summary() string { return "print a small set of system configuration values" }

var getconfValues = map[string]func() string{
	"PAGE_SIZE":         func() string { return "4096" },
	"PAGESIZE":          func() string { return "4096" },
	"_NPROCESSORS_ONLN": func() string { return fmt.Sprintf("%d", runtime.NumCPU()) },
	"ARCH":              unameArch,
}

func (getconfCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) != 1 {
		fmt.Fprintln(ctx.Stderr, "usage: getconf NAME")
		return command.ExitUsage
	}
	fn, ok := getconfValues[ctx.Args[0]]
	if !ok {
		fmt.Fprintf(ctx.Stderr, "getconf: unknown system variable '%s'\n", ctx.Args[0])
		return command.ExitFailure
	}
	fmt.Fprintln(ctx.Stdout, fn())
	return command.ExitSuccess
}

func init() { command.Register(getconfCommand{}) }
