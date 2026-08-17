package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type sortCommand struct{}

func (sortCommand) Name() string    { return "sort" }
func (sortCommand) Summary() string { return "sort lines of text" }

var sortSpec = parser.Spec{
	{Short: 'n', Long: "numeric-sort"},
	{Short: 'r', Long: "reverse"},
	{Short: 'u', Long: "unique"},
	{Short: 'k', Long: "key", HasArg: true},
}

func (sortCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, sortSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "sort: %s\n", err)
		return command.ExitUsage
	}

	numeric := res.Bool('n', "numeric-sort")
	reverse := res.Bool('r', "reverse")
	unique := res.Bool('u', "unique")
	keyField := 0 // 0 means whole line
	if v, ok := res.Value('k', "key"); ok {
		field := v
		if idx := strings.IndexByte(field, ','); idx >= 0 {
			field = field[:idx]
		}
		n, err := strconv.Atoi(field)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "sort: invalid key: '%s'\n", v)
			return command.ExitUsage
		}
		keyField = n
	}

	var lines []string
	exit := command.ExitSuccess

	readAll := func(r io.Reader) error {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		return scanner.Err()
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		if err := readAll(ctx.Stdin); err != nil {
			fmt.Fprintf(ctx.Stderr, "sort: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
	} else {
		for _, arg := range files {
			if arg == "-" {
				if err := readAll(ctx.Stdin); err != nil {
					fmt.Fprintf(ctx.Stderr, "sort: -: %s\n", output.LinuxErrorText(err))
					exit = command.ExitFailure
				}
				continue
			}
			resolved, err := paths.Resolve(arg)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "sort", arg, err)
				exit = command.ExitFailure
				continue
			}
			f, err := os.Open(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "sort", arg, err)
				exit = command.ExitFailure
				continue
			}
			err = readAll(f)
			f.Close()
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "sort", arg, err)
				exit = command.ExitFailure
			}
		}
	}

	keyOf := func(line string) string {
		if keyField <= 0 {
			return line
		}
		fields := strings.Fields(line)
		if keyField > len(fields) {
			return ""
		}
		return fields[keyField-1]
	}

	less := func(i, j int) bool {
		a, b := keyOf(lines[i]), keyOf(lines[j])
		var result bool
		if numeric {
			na, _ := strconv.ParseFloat(strings.TrimSpace(a), 64)
			nb, _ := strconv.ParseFloat(strings.TrimSpace(b), 64)
			result = na < nb
		} else {
			result = a < b
		}
		if reverse {
			return !result && a != b
		}
		return result
	}
	sort.SliceStable(lines, less)

	if unique {
		lines = dedupeAdjacent(lines, keyOf)
	}

	for _, l := range lines {
		fmt.Fprintln(ctx.Stdout, l)
	}
	return exit
}

func dedupeAdjacent(lines []string, keyOf func(string) string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := lines[:1]
	for _, l := range lines[1:] {
		if keyOf(l) == keyOf(out[len(out)-1]) {
			continue
		}
		out = append(out, l)
	}
	return out
}

func init() { command.Register(sortCommand{}) }
