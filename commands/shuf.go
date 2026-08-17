package commands

import (
	"fmt"
	"math/rand"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type shufCommand struct{}

func (shufCommand) Name() string    { return "shuf" }
func (shufCommand) Summary() string { return "randomly permute lines of input" }

var shufSpec = parser.Spec{
	{Short: 'n', HasArg: true},
}

func (shufCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, shufSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "shuf: %s\n", err)
		return command.ExitUsage
	}
	count := -1
	if v, ok := res.Value('n', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "shuf: invalid count '%s'\n", v)
			return command.ExitUsage
		}
		count = n
	}

	var lines []string
	if len(res.Positional) > 0 && res.Positional[0] != "-" {
		lines, err = readAllLines(res.Positional[0], ctx)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "shuf", res.Positional[0], err)
			return command.ExitFailure
		}
	} else {
		lines, err = scanLines(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "shuf: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
	}

	rand.Shuffle(len(lines), func(i, j int) { lines[i], lines[j] = lines[j], lines[i] })
	if count >= 0 && count < len(lines) {
		lines = lines[:count]
	}
	for _, l := range lines {
		fmt.Fprintln(ctx.Stdout, l)
	}
	return command.ExitSuccess
}

func init() { command.Register(shufCommand{}) }
