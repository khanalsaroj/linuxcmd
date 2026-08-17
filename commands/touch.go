package commands

import (
	"fmt"
	"os"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type touchCommand struct{}

func (touchCommand) Name() string    { return "touch" }
func (touchCommand) Summary() string { return "create files / update timestamps" }

var touchSpec = parser.Spec{
	{Short: 'c', Long: "no-create"},
}

func (touchCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, touchSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "touch: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "touch: missing file operand")
		return command.ExitUsage
	}
	noCreate := res.Bool('c', "no-create")

	exit := command.ExitSuccess
	now := time.Now()
	for _, arg := range res.Positional {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "touch", arg, err)
			exit = command.ExitFailure
			continue
		}

		if _, statErr := os.Stat(resolved); statErr != nil {
			if !os.IsNotExist(statErr) {
				output.SimpleErrorf(ctx.Stderr, "touch", arg, statErr)
				exit = command.ExitFailure
				continue
			}
			if noCreate {
				continue
			}
			f, createErr := os.Create(resolved)
			if createErr != nil {
				output.SimpleErrorf(ctx.Stderr, "touch", arg, createErr)
				exit = command.ExitFailure
				continue
			}
			f.Close()
			continue
		}

		if err := os.Chtimes(resolved, now, now); err != nil {
			output.SimpleErrorf(ctx.Stderr, "touch", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(touchCommand{}) }
