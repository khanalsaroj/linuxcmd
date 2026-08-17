package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type helpCommand struct{}

func (helpCommand) Name() string    { return "help" }
func (helpCommand) Summary() string { return "show usage for linuxcmd commands" }

func (helpCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stdout, "Available commands:")
		for _, n := range command.Names() {
			if c, ok := command.Lookup(n); ok {
				fmt.Fprintf(ctx.Stdout, "  %-12s %s\n", n, c.Summary())
			}
		}
		fmt.Fprintln(ctx.Stdout, "\nRun 'help COMMAND' for a specific command.")
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, name := range ctx.Args {
		c, ok := command.Lookup(name)
		if !ok {
			fmt.Fprintf(ctx.Stderr, "help: no help topic for '%s'\n", name)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%s - %s\n", c.Name(), c.Summary())
		fmt.Fprintf(ctx.Stdout, "usage: %s [OPTIONS] [ARGS...]\n", c.Name())
	}
	return exit
}

func init() { command.Register(helpCommand{}) }
