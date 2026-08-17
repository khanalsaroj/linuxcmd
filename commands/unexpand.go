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

type unexpandCommand struct{}

func (unexpandCommand) Name() string    { return "unexpand" }
func (unexpandCommand) Summary() string { return "convert spaces to tabs" }

var unexpandSpec = parser.Spec{
	{Short: 't', HasArg: true},
	{Short: 'a'},
}

// tabifyRun converts a run of `spaces` blanks starting at column startCol
// into the fewest tabs (at each width-column tab stop) plus any leftover
// spaces needed to reach the same column.
func tabifyRun(spaces, startCol, width int) string {
	var b strings.Builder
	col := startCol
	end := startCol + spaces
	for {
		next := (col/width + 1) * width
		if next > end {
			break
		}
		b.WriteByte('\t')
		col = next
	}
	b.WriteString(strings.Repeat(" ", end-col))
	return b.String()
}

// unexpandLine converts runs of spaces to tabs. Without -a, only the
// leading run of spaces is converted (matching GNU unexpand's default);
// with -a, every run of spaces in the line is converted.
func unexpandLine(line string, width int, all bool) string {
	if !all {
		i := 0
		for i < len(line) && line[i] == ' ' {
			i++
		}
		return tabifyRun(i, 0, width) + line[i:]
	}

	var b strings.Builder
	col := 0
	runStart, runLen := 0, 0
	flush := func() {
		if runLen > 0 {
			b.WriteString(tabifyRun(runLen, runStart, width))
			runLen = 0
		}
	}
	for _, r := range line {
		if r == ' ' {
			if runLen == 0 {
				runStart = col
			}
			runLen++
			col++
			continue
		}
		flush()
		b.WriteRune(r)
		col++
	}
	flush()
	return b.String()
}

func (unexpandCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, unexpandSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "unexpand: %s\n", err)
		return command.ExitUsage
	}
	width := 8
	if v, ok := res.Value('t', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "unexpand: invalid tab size '%s'\n", v)
			return command.ExitUsage
		}
		width = n
	}
	all := res.Bool('a', "")

	process := func(r *bufio.Scanner) {
		for r.Scan() {
			fmt.Fprintln(ctx.Stdout, unexpandLine(r.Text(), width, all))
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
			output.SimpleErrorf(ctx.Stderr, "unexpand", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "unexpand", arg, err)
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

func init() { command.Register(unexpandCommand{}) }
