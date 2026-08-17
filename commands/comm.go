package commands

import (
	"fmt"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type commCommand struct{}

func (commCommand) Name() string    { return "comm" }
func (commCommand) Summary() string { return "compare two sorted files line by line" }

var commSpec = parser.Spec{
	{Short: '1'},
	{Short: '2'},
	{Short: '3'},
}

func (commCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, commSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "comm: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: comm [-1] [-2] [-3] FILE1 FILE2")
		return command.ExitUsage
	}
	suppress1 := res.Bool('1', "")
	suppress2 := res.Bool('2', "")
	suppress3 := res.Bool('3', "")

	lines1, err := readAllLines(res.Positional[0], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "comm", res.Positional[0], err)
		return command.ExitFailure
	}
	lines2, err := readAllLines(res.Positional[1], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "comm", res.Positional[1], err)
		return command.ExitFailure
	}

	col2Indent := ""
	if !suppress1 {
		col2Indent = "\t"
	}
	col3Indent := col2Indent
	if !suppress2 {
		col3Indent += "\t"
	}

	i, j := 0, 0
	for i < len(lines1) && j < len(lines2) {
		switch {
		case lines1[i] < lines2[j]:
			if !suppress1 {
				fmt.Fprintln(ctx.Stdout, lines1[i])
			}
			i++
		case lines1[i] > lines2[j]:
			if !suppress2 {
				fmt.Fprintln(ctx.Stdout, col2Indent+lines2[j])
			}
			j++
		default:
			if !suppress3 {
				fmt.Fprintln(ctx.Stdout, col3Indent+lines1[i])
			}
			i++
			j++
		}
	}
	for ; i < len(lines1); i++ {
		if !suppress1 {
			fmt.Fprintln(ctx.Stdout, lines1[i])
		}
	}
	for ; j < len(lines2); j++ {
		if !suppress2 {
			fmt.Fprintln(ctx.Stdout, col2Indent+lines2[j])
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(commCommand{}) }
