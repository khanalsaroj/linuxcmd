package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type tailCommand struct{}

func (tailCommand) Name() string    { return "tail" }
func (tailCommand) Summary() string { return "print the last part of files" }

var tailSpec = parser.Spec{
	{Short: 'n', Long: "lines", HasArg: true},
	{Short: 'c', Long: "bytes", HasArg: true},
	{Short: 'f', Long: "follow"},
	{Short: 'q', Long: "quiet"},
	{Short: 'v', Long: "verbose"},
}

func (tailCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, tailSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tail: %s\n", err)
		return command.ExitUsage
	}

	lines := 10
	bytesN := -1
	if v, ok := res.Value('n', "lines"); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "tail: invalid number of lines: '%s'\n", v)
			return command.ExitUsage
		}
		lines = n
	}
	if v, ok := res.Value('c', "bytes"); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "tail: invalid number of bytes: '%s'\n", v)
			return command.ExitUsage
		}
		bytesN = n
	}
	byteMode := bytesN >= 0
	count := lines
	if byteMode {
		count = bytesN
	}

	follow := res.Bool('f', "follow")
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

	printTail := func(data []byte) {
		if byteMode {
			ctx.Stdout.Write(lastBytes(data, count))
			return
		}
		ctx.Stdout.Write(lastLines(data, count))
	}

	if len(files) == 0 {
		data, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "tail: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		writeHeader("standard input", true)
		printTail(data)
		return command.ExitSuccess
	}

	if follow && len(files) != 1 {
		fmt.Fprintln(ctx.Stderr, "tail: -f only supports a single file")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	var followPath string
	for i, arg := range files {
		if arg == "-" {
			data, err := io.ReadAll(ctx.Stdin)
			if err != nil {
				fmt.Fprintf(ctx.Stderr, "tail: -: %s\n", output.LinuxErrorText(err))
				exit = command.ExitFailure
				continue
			}
			writeHeader("standard input", i == 0)
			printTail(data)
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tail", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tail", arg, err)
			exit = command.ExitFailure
			continue
		}
		if info.IsDir() {
			fmt.Fprintf(ctx.Stderr, "tail: error reading '%s': Is a directory\n", arg)
			exit = command.ExitFailure
			continue
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tail", arg, err)
			exit = command.ExitFailure
			continue
		}
		writeHeader(arg, i == 0)
		printTail(data)
		followPath = resolved
	}

	if follow && exit == command.ExitSuccess && followPath != "" {
		followFile(ctx, followPath)
	}

	return exit
}

// lastLines returns the final n lines of data, preserving whether the
// original ended with a trailing newline.
func lastLines(data []byte, n int) []byte {
	if n <= 0 || len(data) == 0 {
		return nil
	}
	endsWithNewline := data[len(data)-1] == '\n'
	text := data
	if endsWithNewline {
		text = text[:len(text)-1]
	}
	lines := strings.Split(string(text), "\n")
	if n < len(lines) {
		lines = lines[len(lines)-n:]
	}
	joined := strings.Join(lines, "\n")
	if endsWithNewline {
		joined += "\n"
	}
	return []byte(joined)
}

// lastBytes returns the final n bytes of data.
func lastBytes(data []byte, n int) []byte {
	if n >= len(data) {
		return data
	}
	return data[len(data)-n:]
}

// followFile polls path for appended data and streams it to ctx.Stdout,
// the way "tail -f" does. It runs until the process is terminated (e.g.
// Ctrl+C), matching GNU tail's own behavior of never exiting on its own.
func followFile(ctx *command.Context, path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	offset := info.Size()
	for {
		time.Sleep(500 * time.Millisecond)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.Size() < offset {
			// File was truncated/replaced; start over from the beginning.
			offset = 0
		}
		if info.Size() <= offset {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		if _, err := f.Seek(offset, io.SeekStart); err == nil {
			n, _ := io.Copy(ctx.Stdout, f)
			offset += n
		}
		f.Close()
	}
}

func init() { command.Register(tailCommand{}) }
