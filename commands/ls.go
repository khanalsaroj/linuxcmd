package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type lsCommand struct{}

func (lsCommand) Name() string    { return "ls" }
func (lsCommand) Summary() string { return "list directory contents" }

var lsSpec = parser.Spec{
	{Short: 'l'},               // long format
	{Short: 'a', Long: "all"},  // include hidden entries
	{Short: 'h'},               // human-readable sizes (with -l)
	{Short: '1'},               // one entry per line
}

func (lsCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, lsSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ls: %s\n", err)
		return command.ExitUsage
	}

	long := res.Bool('l', "")
	all := res.Bool('a', "all")
	human := res.Bool('h', "")
	one := res.Bool('1', "")

	targets := paths.ExpandGlobs(res.Positional)
	if len(targets) == 0 {
		targets = []string{"."}
	}

	exit := command.ExitSuccess
	multiple := len(targets) > 1
	for i, t := range targets {
		resolved, err := paths.Resolve(t)
		if err != nil {
			output.Errorf(ctx.Stderr, "ls", "cannot access", t, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.Errorf(ctx.Stderr, "ls", "cannot access", t, err)
			exit = command.ExitFailure
			continue
		}

		if multiple {
			if i > 0 {
				fmt.Fprintln(ctx.Stdout)
			}
			fmt.Fprintf(ctx.Stdout, "%s:\n", t)
		}

		if !info.IsDir() {
			printEntries(ctx, []os.FileInfo{info}, long, human, one)
			continue
		}

		entries, err := readDirEntries(resolved, all)
		if err != nil {
			output.Errorf(ctx.Stderr, "ls", "cannot access", t, err)
			exit = command.ExitFailure
			continue
		}
		printEntries(ctx, entries, long, human, one)
	}
	return exit
}

func readDirEntries(dir string, all bool) ([]os.FileInfo, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, de := range des {
		name := de.Name()
		if !all && strings.HasPrefix(name, ".") {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		if !all {
			if hidden, _ := isHiddenAttr(filepath.Join(dir, name)); hidden {
				continue
			}
		}
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool {
		return strings.ToLower(infos[i].Name()) < strings.ToLower(infos[j].Name())
	})
	return infos, nil
}

func printEntries(ctx *command.Context, infos []os.FileInfo, long, human, one bool) {
	if long {
		for _, info := range infos {
			line := output.FormatLongLine(output.LongEntry{Info: info})
			if human {
				line = withHumanSize(line, info)
			}
			fmt.Fprintln(ctx.Stdout, line)
		}
		return
	}

	if one || !isInteractive(ctx) {
		for _, info := range infos {
			fmt.Fprintln(ctx.Stdout, info.Name())
		}
		return
	}

	printColumns(ctx, infos)
}

// withHumanSize re-renders the size field of an already-formatted long
// line using a human-readable unit. Formatting the line first and then
// patching the size keeps FormatLongLine the single source of column
// layout.
func withHumanSize(line string, info os.FileInfo) string {
	full := fmt.Sprintf("%10d", info.Size())
	human := fmt.Sprintf("%10s", output.HumanSize(info.Size()))
	return strings.Replace(line, full, human, 1)
}

func printColumns(ctx *command.Context, infos []os.FileInfo) {
	if len(infos) == 0 {
		return
	}
	const width = 80
	longest := 0
	for _, info := range infos {
		if l := len(info.Name()); l > longest {
			longest = l
		}
	}
	colWidth := longest + 2
	cols := width / colWidth
	if cols < 1 {
		cols = 1
	}
	rows := (len(infos) + cols - 1) / cols
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			idx := c*rows + r
			if idx >= len(infos) {
				continue
			}
			name := infos[idx].Name()
			if c == cols-1 || idx+rows >= len(infos) {
				fmt.Fprint(ctx.Stdout, name)
			} else {
				fmt.Fprint(ctx.Stdout, name+strings.Repeat(" ", colWidth-len(name)))
			}
		}
		fmt.Fprintln(ctx.Stdout)
	}
}

// isInteractive reports whether stdout looks like a real console rather
// than a redirected file/pipe. When redirected, "ls" without -1 still
// prints one entry per line, matching GNU ls' own behavior of disabling
// multi-column output for non-terminal stdout.
func isInteractive(ctx *command.Context) bool {
	f, ok := ctx.Stdout.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

func init() { command.Register(lsCommand{}) }
