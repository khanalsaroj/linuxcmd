package commands

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type expandCommand struct{}

func (expandCommand) Name() string    { return "expand" }
func (expandCommand) Summary() string { return "convert tabs to spaces" }

var expandSpec = parser.Spec{
	{Short: 't', HasArg: true},
}

func expandTabs(line string, width int) string {
	var b strings.Builder
	col := 0
	for _, r := range line {
		if r == '\t' {
			spaces := width - (col % width)
			b.WriteString(strings.Repeat(" ", spaces))
			col += spaces
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String()
}

func (expandCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, expandSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "expand: %s\n", err)
		return command.ExitUsage
	}
	width := 8
	if v, ok := res.Value('t', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "expand: invalid tab size '%s'\n", v)
			return command.ExitUsage
		}
		width = n
	}

	process := func(r *bufio.Scanner) {
		for r.Scan() {
			fmt.Fprintln(ctx.Stdout, expandTabs(r.Text(), width))
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
			output.SimpleErrorf(ctx.Stderr, "expand", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "expand", arg, err)
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

func init() { command.Register(expandCommand{}) }
