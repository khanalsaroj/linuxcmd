package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type duCommand struct{}

func (duCommand) Name() string    { return "du" }
func (duCommand) Summary() string { return "estimate file space usage" }

var duSpec = parser.Spec{
	{Short: 's', Long: "summarize"},
	{Short: 'h', Long: "human-readable"},
}

func (duCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, duSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "du: %s\n", err)
		return command.ExitUsage
	}

	summarize := res.Bool('s', "summarize")
	human := res.Bool('h', "human-readable")

	targets := res.Positional
	if len(targets) == 0 {
		targets = []string{"."}
	}

	format := func(n int64) string {
		if human {
			return output.HumanSize(n)
		}
		return fmt.Sprintf("%d", (n+1023)/1024)
	}

	exit := command.ExitSuccess
	for _, arg := range targets {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "du", arg, err)
			exit = command.ExitFailure
			continue
		}

		info, statErr := os.Stat(resolved)
		if statErr != nil {
			output.SimpleErrorf(ctx.Stderr, "du", arg, statErr)
			exit = command.ExitFailure
			continue
		}
		if !info.IsDir() {
			fmt.Fprintf(ctx.Stdout, "%s\t%s\n", format(info.Size()), arg)
			continue
		}

		sizes := map[string]int64{}
		walkErr := filepath.Walk(resolved, func(p string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			for d := filepath.Dir(p); ; d = filepath.Dir(d) {
				sizes[d] += fi.Size()
				if d == resolved || d == filepath.Dir(d) {
					break
				}
			}
			return nil
		})
		if walkErr != nil {
			output.SimpleErrorf(ctx.Stderr, "du", arg, walkErr)
			exit = command.ExitFailure
			continue
		}

		if summarize {
			fmt.Fprintf(ctx.Stdout, "%s\t%s\n", format(sizes[resolved]), arg)
			continue
		}

		var dirs []string
		for d := range sizes {
			dirs = append(dirs, d)
		}
		sort.Strings(dirs)
		for _, d := range dirs {
			fmt.Fprintf(ctx.Stdout, "%s\t%s\n", format(sizes[d]), d)
		}
	}
	return exit
}

func init() { command.Register(duCommand{}) }
