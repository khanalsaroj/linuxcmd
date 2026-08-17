package commands

import (
	"bufio"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type revCommand struct{}

func (revCommand) Name() string    { return "rev" }
func (revCommand) Summary() string { return "reverse the characters of each line" }

func reverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func (revCommand) Run(ctx *command.Context) int {
	process := func(r *bufio.Scanner) {
		for r.Scan() {
			fmt.Fprintln(ctx.Stdout, reverseRunes(r.Text()))
		}
	}
	newScanner := func(f *os.File) *bufio.Scanner {
		var s *bufio.Scanner
		if f == nil {
			s = bufio.NewScanner(ctx.Stdin)
		} else {
			s = bufio.NewScanner(f)
		}
		s.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		return s
	}

	files := paths.ExpandGlobs(ctx.Args)
	if len(files) == 0 {
		process(newScanner(nil))
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		if arg == "-" {
			process(newScanner(nil))
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "rev", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "rev", arg, err)
			exit = command.ExitFailure
			continue
		}
		process(newScanner(f))
		f.Close()
	}
	return exit
}

func init() { command.Register(revCommand{}) }
