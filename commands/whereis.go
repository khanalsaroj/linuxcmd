package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

type whereisCommand struct{}

func (whereisCommand) Name() string { return "whereis" }
func (whereisCommand) Summary() string {
	return "locate a command's binary and, if registered, its linuxcmd origin"
}

func (whereisCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: whereis COMMAND...")
		return command.ExitUsage
	}

	for _, name := range ctx.Args {
		fmt.Fprintf(ctx.Stdout, "%s:", name)
		if path, ok := findOnPath(name); ok {
			fmt.Fprintf(ctx.Stdout, " %s", path)
		}
		if _, ok := command.Lookup(name); ok {
			fmt.Fprint(ctx.Stdout, " (linuxcmd built-in)")
		}
		fmt.Fprintln(ctx.Stdout)
	}
	return command.ExitSuccess
}

func init() { command.Register(whereisCommand{}) }
