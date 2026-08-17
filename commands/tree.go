package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type treeCommand struct{}

func (treeCommand) Name() string    { return "tree" }
func (treeCommand) Summary() string { return "print a directory tree" }

var treeSpec = parser.Spec{
	{Short: 'L', HasArg: true},
	{Short: 'a', Long: "all"},
}

func (treeCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, treeSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tree: %s\n", err)
		return command.ExitUsage
	}

	maxDepth := -1
	if v, ok := res.Value('L', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fmt.Fprintf(ctx.Stderr, "tree: invalid level: '%s'\n", v)
			return command.ExitUsage
		}
		maxDepth = n
	}
	all := res.Bool('a', "all")

	root := "."
	if len(res.Positional) > 0 {
		root = res.Positional[0]
	}
	resolved, err := paths.Resolve(root)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tree: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	if _, err := os.Stat(resolved); err != nil {
		fmt.Fprintf(ctx.Stderr, "tree: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	fmt.Fprintln(ctx.Stdout, root)
	dirs, files := printTree(ctx, resolved, "", 1, maxDepth, all)
	fmt.Fprintf(ctx.Stdout, "\n%d directories, %d files\n", dirs, files)
	return command.ExitSuccess
}

func printTree(ctx *command.Context, dir, prefix string, depth, maxDepth int, all bool) (dirCount, fileCount int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	var filtered []os.DirEntry
	for _, e := range entries {
		if !all && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		filtered = append(filtered, e)
	}

	for i, e := range filtered {
		last := i == len(filtered)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}
		fmt.Fprintf(ctx.Stdout, "%s%s%s\n", prefix, connector, e.Name())
		if e.IsDir() {
			dirCount++
			if maxDepth < 0 || depth < maxDepth {
				subDirs, subFiles := printTree(ctx, filepath.Join(dir, e.Name()), nextPrefix, depth+1, maxDepth, all)
				dirCount += subDirs
				fileCount += subFiles
			}
		} else {
			fileCount++
		}
	}
	return dirCount, fileCount
}

func init() { command.Register(treeCommand{}) }
