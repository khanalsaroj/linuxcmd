package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type lnCommand struct{}

func (lnCommand) Name() string    { return "ln" }
func (lnCommand) Summary() string { return "create hard or symbolic links" }

var lnSpec = parser.Spec{
	{Short: 's', Long: "symbolic"},
	{Short: 'f', Long: "force"},
}

func (lnCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, lnSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ln: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: ln [-s] [-f] TARGET LINK_NAME")
		return command.ExitUsage
	}
	symbolic := res.Bool('s', "symbolic")
	force := res.Bool('f', "force")

	target := res.Positional[0]
	linkName := res.Positional[1]

	linkResolved, err := paths.Resolve(linkName)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "ln", linkName, err)
		return command.ExitFailure
	}

	if force {
		os.Remove(linkResolved)
	}

	if symbolic {
		// Symlink targets are taken as-is (not resolved against cwd), the
		// way Linux "ln -s" preserves relative targets literally.
		if err := os.Symlink(target, linkResolved); err != nil {
			output.SimpleErrorf(ctx.Stderr, "ln", linkName, err)
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	targetResolved, err := paths.Resolve(target)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "ln", target, err)
		return command.ExitFailure
	}
	if err := os.Link(targetResolved, linkResolved); err != nil {
		output.SimpleErrorf(ctx.Stderr, "ln", linkName, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(lnCommand{}) }
