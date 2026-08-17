package commands

import (
	"strings"

	"linuxcmd/internal/command"
)

type yesCommand struct{}

func (yesCommand) Name() string    { return "yes" }
func (yesCommand) Summary() string { return "repeatedly print a line" }

func (yesCommand) Run(ctx *command.Context) int {
	text := "y"
	if len(ctx.Args) > 0 {
		text = strings.Join(ctx.Args, " ")
	}
	line := text + "\n"

	// Batch many repeats into one buffer before writing, rather than
	// issuing a syscall per line, since this loop is otherwise unbounded.
	const batchLines = 4096
	buf := strings.Repeat(line, batchLines)

	for {
		if _, err := ctx.Stdout.Write([]byte(buf)); err != nil {
			// The reader went away (e.g. piped into "head"); this is the
			// expected way "yes" stops, not a real failure.
			return command.ExitSuccess
		}
	}
}

func init() { command.Register(yesCommand{}) }
