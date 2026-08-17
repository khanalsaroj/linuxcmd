package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type joinCommand struct{}

func (joinCommand) Name() string    { return "join" }
func (joinCommand) Summary() string { return "join lines of two sorted files on a common field" }

func readAllLines(path string, ctx *command.Context) ([]string, error) {
	if path == "-" {
		return scanLines(ctx.Stdin)
	}
	resolved, err := paths.Resolve(path)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(resolved)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return scanLines(f)
}

func scanLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func (joinCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: join FILE1 FILE2")
		return command.ExitUsage
	}

	lines1, err := readAllLines(ctx.Args[0], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "join", ctx.Args[0], err)
		return command.ExitFailure
	}
	lines2, err := readAllLines(ctx.Args[1], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "join", ctx.Args[1], err)
		return command.ExitFailure
	}

	index := make(map[string][]string, len(lines2))
	for _, line := range lines2 {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		index[fields[0]] = append(index[fields[0]], strings.Join(fields[1:], " "))
	}

	for _, line := range lines1 {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		key := fields[0]
		rest1 := strings.Join(fields[1:], " ")
		matches, ok := index[key]
		if !ok {
			continue
		}
		for _, rest2 := range matches {
			out := key
			if rest1 != "" {
				out += " " + rest1
			}
			if rest2 != "" {
				out += " " + rest2
			}
			fmt.Fprintln(ctx.Stdout, out)
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(joinCommand{}) }
