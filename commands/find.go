package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// findCommand implements a useful subset of GNU find: -name, -iname and
// -type predicates over one or more starting paths. find's own syntax
// (single-dash, word-length predicate names like "-name") is not
// POSIX-getopt shaped, so it's hand-parsed here rather than going through
// internal/parser, which is built for combinable short flags.
type findCommand struct{}

func (findCommand) Name() string    { return "find" }
func (findCommand) Summary() string { return "search for files in a directory tree" }

func (findCommand) Run(ctx *command.Context) int {
	var startPaths []string
	var namePattern, inamePattern string
	var typeFilter byte

	args := ctx.Args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-name":
			if i+1 >= len(args) {
				fmt.Fprintln(ctx.Stderr, "find: -name requires an argument")
				return command.ExitUsage
			}
			i++
			namePattern = args[i]
		case "-iname":
			if i+1 >= len(args) {
				fmt.Fprintln(ctx.Stderr, "find: -iname requires an argument")
				return command.ExitUsage
			}
			i++
			inamePattern = args[i]
		case "-type":
			if i+1 >= len(args) {
				fmt.Fprintln(ctx.Stderr, "find: -type requires an argument")
				return command.ExitUsage
			}
			i++
			v := args[i]
			if v != "f" && v != "d" {
				fmt.Fprintf(ctx.Stderr, "find: unsupported -type '%s' (only f, d)\n", v)
				return command.ExitUsage
			}
			typeFilter = v[0]
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(ctx.Stderr, "find: unknown predicate '%s'\n", args[i])
				return command.ExitUsage
			}
			startPaths = append(startPaths, args[i])
		}
	}
	if len(startPaths) == 0 {
		startPaths = []string{"."}
	}

	exit := command.ExitSuccess
	for _, p := range startPaths {
		resolved, err := paths.Resolve(p)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "find", p, err)
			exit = command.ExitFailure
			continue
		}
		if _, err := os.Stat(resolved); err != nil {
			output.SimpleErrorf(ctx.Stderr, "find", p, err)
			exit = command.ExitFailure
			continue
		}

		walkErr := filepath.Walk(resolved, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "find", path, err)
				return nil
			}
			if namePattern != "" {
				if ok, _ := filepath.Match(namePattern, info.Name()); !ok {
					return nil
				}
			}
			if inamePattern != "" {
				if ok, _ := filepath.Match(strings.ToLower(inamePattern), strings.ToLower(info.Name())); !ok {
					return nil
				}
			}
			if typeFilter == 'f' && info.IsDir() {
				return nil
			}
			if typeFilter == 'd' && !info.IsDir() {
				return nil
			}
			fmt.Fprintln(ctx.Stdout, path)
			return nil
		})
		if walkErr != nil {
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() { command.Register(findCommand{}) }
