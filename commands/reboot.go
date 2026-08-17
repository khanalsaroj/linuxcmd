package commands

import "linuxcmd/internal/command"

// rebootCommand is an explicit "shutdown -r now" wrapper.
type rebootCommand struct{}

func (rebootCommand) Name() string    { return "reboot" }
func (rebootCommand) Summary() string { return "restart the machine (shutdown -r now)" }

func (rebootCommand) Run(ctx *command.Context) int {
	return shutdownCommand{}.Run(&command.Context{
		CommandName: "shutdown",
		Args:        []string{"-r", "now"},
		Stdin:       ctx.Stdin,
		Stdout:      ctx.Stdout,
		Stderr:      ctx.Stderr,
	})
}

func init() { command.Register(rebootCommand{}) }
