package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type wcCommand struct{}

func (wcCommand) Name() string    { return "wc" }
func (wcCommand) Summary() string { return "count lines, words, and bytes" }

var wcSpec = parser.Spec{
	{Short: 'l', Long: "lines"},
	{Short: 'w', Long: "words"},
	{Short: 'c', Long: "bytes"},
	{Short: 'm', Long: "chars"},
}

type wcCounts struct {
	lines, words, bytes, chars int64
}

func countStream(r io.Reader) (wcCounts, error) {
	var c wcCounts
	reader := bufio.NewReader(r)
	inWord := false
	buf := make([]byte, 64*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			c.bytes += int64(n)
			c.chars += int64(utf8.RuneCount(chunk))
			for _, b := range chunk {
				if b == '\n' {
					c.lines++
				}
				isSpace := b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' || b == '\f'
				if isSpace {
					inWord = false
				} else if !inWord {
					inWord = true
					c.words++
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func (wcCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, wcSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wc: %s\n", err)
		return command.ExitUsage
	}

	showLines := res.Bool('l', "lines")
	showWords := res.Bool('w', "words")
	showBytes := res.Bool('c', "bytes")
	showChars := res.Bool('m', "chars")
	if !showLines && !showWords && !showBytes && !showChars {
		showLines, showWords, showBytes = true, true, true
	}

	print := func(name string, c wcCounts, showName bool) {
		if showLines {
			fmt.Fprintf(ctx.Stdout, "%8d", c.lines)
		}
		if showWords {
			fmt.Fprintf(ctx.Stdout, "%8d", c.words)
		}
		if showChars {
			fmt.Fprintf(ctx.Stdout, "%8d", c.chars)
		}
		if showBytes {
			fmt.Fprintf(ctx.Stdout, "%8d", c.bytes)
		}
		if showName {
			fmt.Fprintf(ctx.Stdout, " %s", name)
		}
		fmt.Fprintln(ctx.Stdout)
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		c, err := countStream(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "wc: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		print("", c, false)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	var total wcCounts
	for _, arg := range files {
		if arg == "-" {
			c, err := countStream(ctx.Stdin)
			if err != nil {
				fmt.Fprintf(ctx.Stderr, "wc: -: %s\n", output.LinuxErrorText(err))
				exit = command.ExitFailure
				continue
			}
			print("-", c, true)
			total.lines += c.lines
			total.words += c.words
			total.bytes += c.bytes
			total.chars += c.chars
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "wc", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "wc", arg, err)
			exit = command.ExitFailure
			continue
		}
		c, err := countStream(f)
		f.Close()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "wc", arg, err)
			exit = command.ExitFailure
			continue
		}
		print(arg, c, true)
		total.lines += c.lines
		total.words += c.words
		total.bytes += c.bytes
		total.chars += c.chars
	}
	if len(files) > 1 {
		print("total", total, true)
	}
	return exit
}

func init() { command.Register(wcCommand{}) }
