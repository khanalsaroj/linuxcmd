package commands

import (
	"fmt"
	"os"
	"runtime"

	"linuxcmd/internal/command"
)

type lscpuCommand struct{}

func (lscpuCommand) Name() string    { return "lscpu" }
func (lscpuCommand) Summary() string { return "display CPU architecture information" }

func (lscpuCommand) Run(ctx *command.Context) int {
	identifier := os.Getenv("PROCESSOR_IDENTIFIER")
	if identifier == "" {
		identifier = "unknown"
	}
	fmt.Fprintf(ctx.Stdout, "Architecture:        %s\n", unameArch())
	fmt.Fprintf(ctx.Stdout, "CPU(s):              %d\n", runtime.NumCPU())
	fmt.Fprintf(ctx.Stdout, "Model name:          %s\n", identifier)
	if rev := os.Getenv("PROCESSOR_REVISION"); rev != "" {
		fmt.Fprintf(ctx.Stdout, "Revision:            %s\n", rev)
	}
	return command.ExitSuccess
}

func init() { command.Register(lscpuCommand{}) }
