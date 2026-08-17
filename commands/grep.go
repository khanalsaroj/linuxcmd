package commands

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type grepCommand struct{}

func (grepCommand) Name() string    { return "grep" }
func (grepCommand) Summary() string { return "search text using patterns" }

var grepSpec = parser.Spec{
	{Short: 'i', Long: "ignore-case"},
	{Short: 'n', Long: "line-number"},
	{Short: 'v', Long: "invert-match"},
	{Short: 'r'},
	{Short: 'R'},
	{Short: 'l', Long: "files-with-matches"},
	{Short: 'c', Long: "count"},
}

func (grepCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, grepSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "grep: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: grep [OPTION]... PATTERN [FILE]...")
		return command.ExitUsage
	}

	pattern := res.Positional[0]
	files := paths.ExpandGlobs(res.Positional[1:])

	ignoreCase := res.Bool('i', "ignore-case")
	lineNumbers := res.Bool('n', "line-number")
	invert := res.Bool('v', "invert-match")
	recursive := res.Bool('r', "") || res.Bool('R', "")
	listOnly := res.Bool('l', "files-with-matches")
	countOnly := res.Bool('c', "count")

	reSrc := pattern
	if ignoreCase {
		reSrc = "(?i)" + reSrc
	}
	re, err := regexp.Compile(reSrc)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "grep: invalid pattern: %s\n", err)
		return command.ExitUsage
	}

	matchedAny := false
	hadError := false

	process := func(name string, r io.Reader, showName bool) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		lineNo := 0
		count := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			isMatch := re.MatchString(line)
			if invert {
				isMatch = !isMatch
			}
			if !isMatch {
				continue
			}
			matchedAny = true
			count++
			if listOnly {
				break
			}
			if countOnly {
				continue
			}
			prefix := ""
			if showName {
				prefix += name + ":"
			}
			if lineNumbers {
				prefix += fmt.Sprintf("%d:", lineNo)
			}
			fmt.Fprintf(ctx.Stdout, "%s%s\n", prefix, line)
		}
		if listOnly && count > 0 {
			fmt.Fprintln(ctx.Stdout, name)
		}
		if countOnly {
			if showName {
				fmt.Fprintf(ctx.Stdout, "%s:%d\n", name, count)
			} else {
				fmt.Fprintf(ctx.Stdout, "%d\n", count)
			}
		}
	}

	switch {
	case len(files) == 0:
		process("(standard input)", ctx.Stdin, false)

	case recursive:
		for _, f := range files {
			resolved, err := paths.Resolve(f)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "grep", f, err)
				hadError = true
				continue
			}
			walkErr := filepath.Walk(resolved, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					output.SimpleErrorf(ctx.Stderr, "grep", path, err)
					hadError = true
					return nil
				}
				if info.IsDir() {
					return nil
				}
				file, err := os.Open(path)
				if err != nil {
					output.SimpleErrorf(ctx.Stderr, "grep", path, err)
					hadError = true
					return nil
				}
				process(path, file, true)
				file.Close()
				return nil
			})
			if walkErr != nil {
				hadError = true
			}
		}

	default:
		showName := len(files) > 1
		for _, f := range files {
			resolved, err := paths.Resolve(f)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "grep", f, err)
				hadError = true
				continue
			}
			file, err := os.Open(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "grep", f, err)
				hadError = true
				continue
			}
			process(f, file, showName)
			file.Close()
		}
	}

	if hadError {
		return 2
	}
	if matchedAny {
		return command.ExitSuccess
	}
	return command.ExitFailure
}

func init() { command.Register(grepCommand{}) }
