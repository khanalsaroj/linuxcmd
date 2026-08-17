package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type catCommand struct{}

func (catCommand) Name() string    { return "cat" }
func (catCommand) Summary() string { return "concatenate and print files" }

var catSpec = parser.Spec{
	{Short: 'n', Long: "number"},
}

func (catCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, catSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cat: %s\n", err)
		return command.ExitUsage
	}

	numbered := res.Bool('n', "number")
	lineNo := 1

	write := func(r io.Reader) error {
		if !numbered {
			_, err := io.Copy(ctx.Stdout, r)
			return err
		}
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			fmt.Fprintf(ctx.Stdout, "%6d\t%s\n", lineNo, scanner.Text())
			lineNo++
		}
		return scanner.Err()
	}

	if len(res.Positional) == 0 {
		if err := write(ctx.Stdin); err != nil {
			fmt.Fprintf(ctx.Stderr, "cat: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(res.Positional) {
		if arg == "-" {
			if err := write(ctx.Stdin); err != nil {
				fmt.Fprintf(ctx.Stderr, "cat: -: %s\n", output.LinuxErrorText(err))
				exit = command.ExitFailure
			}
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cat", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cat", arg, err)
			exit = command.ExitFailure
			continue
		}
		if info.IsDir() {
			fmt.Fprintf(ctx.Stderr, "cat: %s: Is a directory\n", arg)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cat", arg, err)
			exit = command.ExitFailure
			continue
		}
		err = write(f)
		f.Close()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cat", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(catCommand{}) }
