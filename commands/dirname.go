package commands

import (
	"fmt"
	"path/filepath"

	"linuxcmd/internal/command"
)

type dirnameCommand struct{}

func (dirnameCommand) Name() string    { return "dirname" }
func (dirnameCommand) Summary() string { return "print the directory component of a path" }

func (dirnameCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: dirname PATH")
		return command.ExitUsage
	}
	for _, arg := range ctx.Args {
		fmt.Fprintln(ctx.Stdout, filepath.Dir(filepath.FromSlash(arg)))
	}
	return command.ExitSuccess
}

func init() { command.Register(dirnameCommand{}) }
