package commands

import "linuxcmd/internal/command"

// poweroffCommand is an explicit "shutdown -h now" wrapper.
type poweroffCommand struct{}

func (poweroffCommand) Name() string    { return "poweroff" }
func (poweroffCommand) Summary() string { return "shut the machine down (shutdown -h now)" }

func (poweroffCommand) Run(ctx *command.Context) int {
	return shutdownCommand{}.Run(&command.Context{
		CommandName: "shutdown",
		Args:        []string{"-h", "now"},
		Stdin:       ctx.Stdin,
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
	})
}

func init() { command.Register(poweroffCommand{}) }
