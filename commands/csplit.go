package commands

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type csplitCommand struct{}

func (csplitCommand) Name() string    { return "csplit" }
func (csplitCommand) Summary() string { return "split a file at line-number or regex boundaries" }

func (csplitCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: csplit FILE PATTERN...")
		return command.ExitUsage
	}
	lines, err := readAllLines(ctx.Args[0], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "csplit", ctx.Args[0], err)
		return command.ExitFailure
	}

	var boundaries []int
	pos := 0
	for _, pat := range ctx.Args[1:] {
		if n, err := strconv.Atoi(pat); err == nil {
			idx := n - 1
			boundaries = append(boundaries, idx)
			pos = idx
			continue
		}
		if strings.HasPrefix(pat, "/") && strings.HasSuffix(pat, "/") && len(pat) >= 2 {
			re, err := regexp.Compile(pat[1 : len(pat)-1])
			if err != nil {
				fmt.Fprintf(ctx.Stderr, "csplit: invalid pattern '%s'\n", pat)
				return command.ExitUsage
			}
			found := -1
			for i := pos; i < len(lines); i++ {
				if re.MatchString(lines[i]) {
					found = i
					break
				}
			}
			if found < 0 {
				fmt.Fprintf(ctx.Stderr, "csplit: no match for '%s'\n", pat)
				return command.ExitFailure
			}
			boundaries = append(boundaries, found)
			pos = found
			continue
		}
		fmt.Fprintf(ctx.Stderr, "csplit: unrecognized pattern '%s'\n", pat)
		return command.ExitUsage
	}

	writeChunk := func(chunkLines []string, part int) error {
		name := fmt.Sprintf("xx%02d", part)
		resolved, err := paths.Resolve(name)
		if err != nil {
			return err
		}
		f, err := os.Create(resolved)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, l := range chunkLines {
			fmt.Fprintln(f, l)
		}
		fmt.Fprintln(ctx.Stdout, name)
		return nil
	}

	chunkStart := 0
	part := 0
	for _, b := range boundaries {
		if b <= chunkStart || b > len(lines) {
			continue
		}
		if err := writeChunk(lines[chunkStart:b], part); err != nil {
			fmt.Fprintf(ctx.Stderr, "csplit: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		part++
		chunkStart = b
	}
	if err := writeChunk(lines[chunkStart:], part); err != nil {
		fmt.Fprintf(ctx.Stderr, "csplit: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(csplitCommand{}) }
