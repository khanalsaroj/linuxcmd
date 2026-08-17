package commands

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type columnCommand struct{}

func (columnCommand) Name() string    { return "column" }
func (columnCommand) Summary() string { return "align delimiter-separated input into columns" }

var columnSpec = parser.Spec{
	{Short: 't'},
	{Short: 's', HasArg: true},
}

var columnDefaultSplit = regexp.MustCompile(`\s+`)

func (columnCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, columnSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "column: %s\n", err)
		return command.ExitUsage
	}
	sep, hasSep := res.Value('s', "")

	scanner := bufio.NewScanner(ctx.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var rows [][]string
	widest := 0
	for scanner.Scan() {
		line := scanner.Text()
		var fields []string
		if hasSep {
			fields = strings.Split(line, sep)
		} else {
			fields = columnDefaultSplit.Split(strings.TrimSpace(line), -1)
		}
		rows = append(rows, fields)
		if len(fields) > widest {
			widest = len(fields)
		}
	}

	colWidths := make([]int, widest)
	for _, row := range rows {
		for i, f := range row {
			if len(f) > colWidths[i] {
				colWidths[i] = len(f)
			}
		}
	}

	for _, row := range rows {
		for i, f := range row {
			if i == len(row)-1 {
				fmt.Fprint(ctx.Stdout, f)
				continue
			}
			fmt.Fprintf(ctx.Stdout, "%-*s  ", colWidths[i], f)
		}
		fmt.Fprintln(ctx.Stdout)
	}
	return command.ExitSuccess
}

func init() { command.Register(columnCommand{}) }
