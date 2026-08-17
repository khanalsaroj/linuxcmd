package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type headCommand struct{}

func (headCommand) Name() string    { return "head" }
func (headCommand) Summary() string { return "print the first part of files" }

var headSpec = parser.Spec{
	{Short: 'n', Long: "lines", HasArg: true},
	{Short: 'c', Long: "bytes", HasArg: true},
	{Short: 'q', Long: "quiet"},
	{Short: 'v', Long: "verbose"},
}

func (headCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, headSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "head: %s\n", err)
		return command.ExitUsage
	}

	lines := 10
	bytesN := -1
	if v, ok := res.Value('n', "lines"); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "head: invalid number of lines: '%s'\n", v)
			return command.ExitUsage
		}
		lines = n
	}
	if v, ok := res.Value('c', "bytes"); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "head: invalid number of bytes: '%s'\n", v)
			return command.ExitUsage
		}
		bytesN = n
	}
	byteMode := bytesN >= 0
	count := lines
	if byteMode {
		count = bytesN
	}

	quiet := res.Bool('q', "quiet")
	verbose := res.Bool('v', "verbose")

	files := paths.ExpandGlobs(res.Positional)
	showHeaders := (len(files) > 1 || verbose) && !quiet

	writeHeader := func(name string, first bool) {
		if !showHeaders {
			return
		}
		if !first {
			fmt.Fprintln(ctx.Stdout)
		}
		fmt.Fprintf(ctx.Stdout, "==> %s <==\n", name)
	}

	process := func(r io.Reader) error {
		if byteMode {
			_, err := io.CopyN(ctx.Stdout, r, int64(count))
			if err == io.EOF {
				return nil
			}
			return err
		}
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for i := 0; i < count && scanner.Scan(); i++ {
			fmt.Fprintln(ctx.Stdout, scanner.Text())
		}
		return scanner.Err()
	}

	if len(files) == 0 {
		writeHeader("standard input", true)
		if err := process(ctx.Stdin); err != nil {
			fmt.Fprintf(ctx.Stderr, "head: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for i, arg := range files {
		if arg == "-" {
			writeHeader("standard input", i == 0)
			if err := process(ctx.Stdin); err != nil {
				fmt.Fprintf(ctx.Stderr, "head: -: %s\n", output.LinuxErrorText(err))
				exit = command.ExitFailure
			}
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "head", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "head", arg, err)
			exit = command.ExitFailure
			continue
		}
		if info.IsDir() {
			fmt.Fprintf(ctx.Stderr, "head: error reading '%s': Is a directory\n", arg)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "head", arg, err)
			exit = command.ExitFailure
			continue
		}
		writeHeader(arg, i == 0)
		err = process(f)
		f.Close()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "head", arg, err)
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(headCommand{}) }
