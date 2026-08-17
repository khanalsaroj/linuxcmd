package commands

import (
	"fmt"
	"os"
	"regexp"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

// xRun matches one or more consecutive X characters, the placeholder
// mktemp templates use for the random portion of the name.
var xRun = regexp.MustCompile("X+")

type mktempCommand struct{}

func (mktempCommand) Name() string    { return "mktemp" }
func (mktempCommand) Summary() string { return "create a unique temporary file or directory" }

var mktempSpec = parser.Spec{
	{Short: 'd', Long: "directory"},
}

func (mktempCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, mktempSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mktemp: %s\n", err)
		return command.ExitUsage
	}
	directory := res.Bool('d', "directory")

	template := "tmp.XXXXXXXXXX"
	if len(res.Positional) > 0 {
		template = res.Positional[0]
	}
	pattern := xRun.ReplaceAllString(template, "*")
	if pattern == template {
		pattern += "*"
	}

	if directory {
		name, err := os.MkdirTemp("", pattern)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "mktemp: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		fmt.Fprintln(ctx.Stdout, name)
		return command.ExitSuccess
	}

	f, err := os.CreateTemp("", pattern)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mktemp: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	name := f.Name()
	f.Close()
	fmt.Fprintln(ctx.Stdout, name)
	return command.ExitSuccess
}

func init() { command.Register(mktempCommand{}) }
