package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type cutCommand struct{}

func (cutCommand) Name() string    { return "cut" }
func (cutCommand) Summary() string { return "select fields or columns from each line" }

var cutSpec = parser.Spec{
	{Short: 'd', Long: "delimiter", HasArg: true},
	{Short: 'f', Long: "fields", HasArg: true},
	{Short: 'c', Long: "characters", HasArg: true},
}

// cutRange is an inclusive 1-based [start, end] range; end == -1 means "to
// the end of the line", matching cut's "N-" syntax.
type cutRange struct{ start, end int }

func parseRanges(spec string) ([]cutRange, error) {
	var ranges []cutRange
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.IndexByte(part, '-'); idx >= 0 {
			startStr, endStr := part[:idx], part[idx+1:]
			start := 1
			if startStr != "" {
				n, err := strconv.Atoi(startStr)
				if err != nil {
					return nil, fmt.Errorf("invalid range: '%s'", part)
				}
				start = n
			}
			end := -1
			if endStr != "" {
				n, err := strconv.Atoi(endStr)
				if err != nil {
					return nil, fmt.Errorf("invalid range: '%s'", part)
				}
				end = n
			}
			ranges = append(ranges, cutRange{start, end})
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid field: '%s'", part)
		}
		ranges = append(ranges, cutRange{n, n})
	}
	return ranges, nil
}

func inRanges(ranges []cutRange, i int) bool {
	for _, r := range ranges {
		if i >= r.start && (r.end == -1 || i <= r.end) {
			return true
		}
	}
	return false
}

func (cutCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, cutSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cut: %s\n", err)
		return command.ExitUsage
	}

	delim := "\t"
	if v, ok := res.Value('d', "delimiter"); ok {
		delim = v
	}

	fieldsSpec, byFields := res.Value('f', "fields")
	charsSpec, byChars := res.Value('c', "characters")
	if !byFields && !byChars {
		fmt.Fprintln(ctx.Stderr, "cut: you must specify a list of -f fields or -c characters")
		return command.ExitUsage
	}

	var ranges []cutRange
	if byFields {
		ranges, err = parseRanges(fieldsSpec)
	} else {
		ranges, err = parseRanges(charsSpec)
	}
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "cut: %s\n", err)
		return command.ExitUsage
	}

	process := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if byChars {
				runes := []rune(line)
				var out []rune
				for i, ch := range runes {
					if inRanges(ranges, i+1) {
						out = append(out, ch)
					}
				}
				fmt.Fprintln(ctx.Stdout, string(out))
				continue
			}
			parts := strings.Split(line, delim)
			var selected []string
			for i, p := range parts {
				if inRanges(ranges, i+1) {
					selected = append(selected, p)
				}
			}
			fmt.Fprintln(ctx.Stdout, strings.Join(selected, delim))
		}
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		process(ctx.Stdin)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		if arg == "-" {
			process(ctx.Stdin)
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cut", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cut", arg, err)
			exit = command.ExitFailure
			continue
		}
		process(f)
		f.Close()
	}
	return exit
}

func init() { command.Register(cutCommand{}) }
