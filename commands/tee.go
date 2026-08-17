package commands

import (
	"fmt"
	"io"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type teeCommand struct{}

func (teeCommand) Name() string    { return "tee" }
func (teeCommand) Summary() string { return "copy stdin to stdout and files" }

var teeSpec = parser.Spec{
	{Short: 'a', Long: "append"},
}

func (teeCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, teeSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tee: %s\n", err)
		return command.ExitUsage
	}

	append_ := res.Bool('a', "append")
	flags := os.O_WRONLY | os.O_CREATE
	if append_ {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	writers := []io.Writer{ctx.Stdout}
	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(res.Positional) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tee", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.OpenFile(resolved, flags, 0644)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tee", arg, err)
			exit = command.ExitFailure
			continue
		}
		defer f.Close()
		writers = append(writers, f)
	}

	mw := io.MultiWriter(writers...)
	if _, err := io.Copy(mw, ctx.Stdin); err != nil {
		fmt.Fprintf(ctx.Stderr, "tee: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	return exit
}

func init() { command.Register(teeCommand{}) }
