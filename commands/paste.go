package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type pasteCommand struct{}

func (pasteCommand) Name() string    { return "paste" }
func (pasteCommand) Summary() string { return "merge corresponding lines of files" }

var pasteSpec = parser.Spec{
	{Short: 'd', HasArg: true},
}

func (pasteCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, pasteSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "paste: %s\n", err)
		return command.ExitUsage
	}
	delim := "\t"
	if v, ok := res.Value('d', ""); ok {
		delim = v
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: paste [-d DELIM] FILE...")
		return command.ExitUsage
	}

	var scanners []*bufio.Scanner
	var closers []*os.File
	defer func() {
		for _, f := range closers {
			f.Close()
		}
	}()

	exit := command.ExitSuccess
	for _, arg := range files {
		if arg == "-" {
			s := bufio.NewScanner(ctx.Stdin)
			s.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
			scanners = append(scanners, s)
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "paste", arg, err)
			exit = command.ExitFailure
			scanners = append(scanners, nil)
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "paste", arg, err)
			exit = command.ExitFailure
			scanners = append(scanners, nil)
			continue
		}
		closers = append(closers, f)
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		scanners = append(scanners, s)
	}

	for {
		anyMore := false
		var fields []string
		for _, s := range scanners {
			if s != nil && s.Scan() {
				fields = append(fields, s.Text())
				anyMore = true
			} else {
				fields = append(fields, "")
			}
		}
		if !anyMore {
			break
		}
		fmt.Fprintln(ctx.Stdout, strings.Join(fields, delim))
	}
	return exit
}

func init() { command.Register(pasteCommand{}) }
