package commands

import (
	"bufio"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type cmpCommand struct{}

func (cmpCommand) Name() string    { return "cmp" }
func (cmpCommand) Summary() string { return "compare two files byte by byte" }

func (cmpCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: cmp FILE1 FILE2")
		return command.ExitUsage
	}

	open := func(name string) (*bufio.Reader, func(), error) {
		if name == "-" {
			return bufio.NewReader(ctx.Stdin), func() {}, nil
		}
		resolved, err := paths.Resolve(name)
		if err != nil {
			return nil, nil, err
		}
		f, err := os.Open(resolved)
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewReader(f), func() { f.Close() }, nil
	}

	r1, close1, err := open(ctx.Args[0])
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "cmp", ctx.Args[0], err)
		return command.ExitFailure
	}
	defer close1()
	r2, close2, err := open(ctx.Args[1])
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "cmp", ctx.Args[1], err)
		return command.ExitFailure
	}
	defer close2()

	byteNum := int64(0)
	line := int64(1)
	for {
		b1, err1 := r1.ReadByte()
		b2, err2 := r2.ReadByte()
		byteNum++

		if err1 != nil && err2 != nil {
			return command.ExitSuccess
		}
		if err1 != nil {
			fmt.Fprintf(ctx.Stdout, "cmp: EOF on %s\n", ctx.Args[0])
			return command.ExitFailure
		}
		if err2 != nil {
			fmt.Fprintf(ctx.Stdout, "cmp: EOF on %s\n", ctx.Args[1])
			return command.ExitFailure
		}
		if b1 == '\n' {
			line++
		}
		if b1 != b2 {
			fmt.Fprintf(ctx.Stdout, "%s %s differ: byte %d, line %d\n", ctx.Args[0], ctx.Args[1], byteNum, line)
			return command.ExitFailure
		}
	}
}

func init() { command.Register(cmpCommand{}) }
