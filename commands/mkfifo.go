package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

// mkfifoCommand reports the operation as unsupported: native Windows has
// no filesystem-visible FIFO/named-pipe object comparable to a POSIX
// FIFO file (Windows named pipes live in the \\.\pipe\ namespace, not on
// disk, and have different semantics), so faking one would be
// misleading. This is the "report unsupported" option the backlog itself
// calls out as acceptable.
type mkfifoCommand struct{}

func (mkfifoCommand) Name() string    { return "mkfifo" }
func (mkfifoCommand) Summary() string { return "create a FIFO (unsupported on Windows)" }

func (mkfifoCommand) Run(ctx *command.Context) int {
	fmt.Fprintln(ctx.Stderr, "mkfifo: not supported on native Windows (no on-disk FIFO object); "+
		"see \\\\.\\pipe\\ named pipes for the closest Windows equivalent")
	return command.ExitFailure
}

func init() { command.Register(mkfifoCommand{}) }
