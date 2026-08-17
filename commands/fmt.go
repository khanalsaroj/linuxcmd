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

type fmtCommand struct{}

func (fmtCommand) Name() string    { return "fmt" }
func (fmtCommand) Summary() string { return "reflow plain-text paragraphs" }

var fmtSpec = parser.Spec{
	{Short: 'w', HasArg: true},
}

// wrapParagraph greedily fills lines up to width, the way GNU fmt does
// for plain (non-hyphenated) text.
func wrapParagraph(words []string, width int) []string {
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur strings.Builder
	curLen := 0
	for _, w := range words {
		wLen := len([]rune(w))
		if curLen == 0 {
			cur.WriteString(w)
			curLen = wLen
			continue
		}
		if curLen+1+wLen > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
			curLen = wLen
			continue
		}
		cur.WriteByte(' ')
		cur.WriteString(w)
		curLen += 1 + wLen
	}
	if curLen > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

func (fmtCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, fmtSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "fmt: %s\n", err)
		return command.ExitUsage
	}
	width := 75
	if v, ok := res.Value('w', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "fmt: invalid width '%s'\n", v)
			return command.ExitUsage
		}
		width = n
	}

	process := func(r *bufio.Scanner) {
		var para []string
		flush := func() {
			for _, line := range wrapParagraph(para, width) {
				fmt.Fprintln(ctx.Stdout, line)
			}
			para = nil
		}
		for r.Scan() {
			line := strings.TrimSpace(r.Text())
			if line == "" {
				flush()
				fmt.Fprintln(ctx.Stdout)
				continue
			}
			para = append(para, strings.Fields(line)...)
		}
		flush()
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
			output.SimpleErrorf(ctx.Stderr, "fmt", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "fmt", arg, err)
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

func init() { command.Register(fmtCommand{}) }
