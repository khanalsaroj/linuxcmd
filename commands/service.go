package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

// serviceCommand is "service NAME VERB" instead of systemctl's
// "systemctl VERB NAME" argument order; both wrap the same sc.exe calls.
type serviceCommand struct{}

func (serviceCommand) Name() string    { return "service" }
func (serviceCommand) Summary() string { return "control a Windows service (NAME status/start/stop)" }

func (serviceCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: service SERVICE {status|start|stop|restart}")
		return command.ExitUsage
	}
	name, verb := ctx.Args[0], ctx.Args[1]
	return systemctlCommand{}.Run(&command.Context{
		CommandName: "systemctl",
		Args:        []string{verb, name},
		Stdin:       ctx.Stdin,
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
	})
}

func init() { command.Register(serviceCommand{}) }
