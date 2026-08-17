package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type niceCommand struct{}

func (niceCommand) Name() string    { return "nice" }
func (niceCommand) Summary() string { return "run a command with an adjusted scheduling priority" }

var niceSpec = parser.Spec{
	{Short: 'n', HasArg: true},
}

// Win32 process priority class values (winbase.h). Not exposed by the
// standard syscall package, so defined here directly.
const (
	idlePriorityClass        = 0x00000040
	belowNormalPriorityClass = 0x00004000
	normalPriorityClass      = 0x00000020
	aboveNormalPriorityClass = 0x00008000
	highPriorityClass        = 0x00000080
)

// niceToPriorityClass maps a Linux nice value (-20 highest .. 19 lowest)
// onto the nearest Windows priority class, since Windows has no
// equivalent numeric scale.
func niceToPriorityClass(n int) uint32 {
	switch {
	case n <= -10:
		return highPriorityClass
	case n < 0:
		return aboveNormalPriorityClass
	case n == 0:
		return normalPriorityClass
	case n < 10:
		return belowNormalPriorityClass
	default:
		return idlePriorityClass
	}
}

func (niceCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, niceSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "nice: %s\n", err)
		return command.ExitUsage
	}
	niceVal := 10
	if v, ok := res.Value('n', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "nice: invalid adjustment '%s'\n", v)
			return command.ExitUsage
		}
		niceVal = n
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: nice [-n ADJUSTMENT] COMMAND [ARG...]")
		return command.ExitUsage
	}

	cmd := exec.Command(res.Positional[0], res.Positional[1:]...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: niceToPriorityClass(niceVal)}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "nice: %s\n", err)
		return command.ExitNotFound
	}
	return command.ExitSuccess
}

func init() { command.Register(niceCommand{}) }
