package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
)

type basenameCommand struct{}

func (basenameCommand) Name() string    { return "basename" }
func (basenameCommand) Summary() string { return "strip directory and suffix from a path" }

func (basenameCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: basename PATH [SUFFIX]")
		return command.ExitUsage
	}
	name := filepath.Base(filepath.FromSlash(ctx.Args[0]))
	if len(ctx.Args) >= 2 && ctx.Args[1] != "" && name != ctx.Args[1] {
		name = strings.TrimSuffix(name, ctx.Args[1])
	}
	fmt.Fprintln(ctx.Stdout, name)
	return command.ExitSuccess
}

func init() { command.Register(basenameCommand{}) }
