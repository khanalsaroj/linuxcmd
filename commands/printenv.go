package commands

import (
	"fmt"
	"os"
	"sort"

	"linuxcmd/internal/command"
)

type printenvCommand struct{}

func (printenvCommand) Name() string    { return "printenv" }
func (printenvCommand) Summary() string { return "print environment variables" }

func (printenvCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		env := os.Environ()
		sort.Strings(env)
		for _, kv := range env {
			fmt.Fprintln(ctx.Stdout, kv)
		}
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, name := range ctx.Args {
		v, ok := os.LookupEnv(name)
		if !ok {
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintln(ctx.Stdout, v)
	}
	return exit
}

func init() { command.Register(printenvCommand{}) }
