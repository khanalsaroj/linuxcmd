package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type foldCommand struct{}

func (foldCommand) Name() string    { return "fold" }
func (foldCommand) Summary() string { return "wrap each line to a given width" }

var foldSpec = parser.Spec{
	{Short: 'w', HasArg: true},
}

func foldLine(line string, width int) []string {
	runes := []rune(line)
	if len(runes) == 0 {
		return []string{""}
	}
	var out []string
	for len(runes) > width {
		out = append(out, string(runes[:width]))
		runes = runes[width:]
	}
	out = append(out, string(runes))
	return out
}

func (foldCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, foldSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "fold: %s\n", err)
		return command.ExitUsage
	}
	width := 80
	if v, ok := res.Value('w', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "fold: invalid width '%s'\n", v)
			return command.ExitUsage
		}
		width = n
	}

	process := func(r *bufio.Scanner) {
		for r.Scan() {
			for _, part := range foldLine(r.Text(), width) {
				fmt.Fprintln(ctx.Stdout, part)
			}
		}
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		scanner := bufio.NewScanner(ctx.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "fold", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "fold", arg, err)
			exit = command.ExitFailure
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		f.Close()
	}
	return exit
}

func init() { command.Register(foldCommand{}) }
