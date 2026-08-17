package commands

import (
	"bufio"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type tacCommand struct{}

func (tacCommand) Name() string    { return "tac" }
func (tacCommand) Summary() string { return "print lines in reverse order" }

func (tacCommand) Run(ctx *command.Context) int {
	readLines := func(r *bufio.Scanner) []string {
		var lines []string
		for r.Scan() {
			lines = append(lines, r.Text())
		}
		return lines
	}

	files := paths.ExpandGlobs(ctx.Args)
	var lines []string
	exit := command.ExitSuccess

	if len(files) == 0 {
		scanner := bufio.NewScanner(ctx.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		lines = readLines(scanner)
	} else {
		for _, arg := range files {
			var f *os.File
			if arg == "-" {
				scanner := bufio.NewScanner(ctx.Stdin)
				scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
				lines = append(lines, readLines(scanner)...)
				continue
			}
			resolved, err := paths.Resolve(arg)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "tac", arg, err)
				exit = command.ExitFailure
				continue
			}
			f, err = os.Open(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "tac", arg, err)
				exit = command.ExitFailure
				continue
			}
			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
			lines = append(lines, readLines(scanner)...)
			f.Close()
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		fmt.Fprintln(ctx.Stdout, lines[i])
	}
	return exit
}

func init() { command.Register(tacCommand{}) }
