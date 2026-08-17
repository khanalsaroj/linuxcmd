package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
)

type whichCommand struct{}

func (whichCommand) Name() string    { return "which" }
func (whichCommand) Summary() string { return "locate a command on PATH" }

func (whichCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: which COMMAND...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, name := range ctx.Args {
		found, ok := findOnPath(name)
		if !ok {
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintln(ctx.Stdout, found)
	}
	return exit
}

// findOnPath mimics how CMD resolves a bare command name: it tries each
// PATH directory, and within each directory tries the name as-is (if it
// already has an extension) and with each PATHEXT suffix in order.
func findOnPath(name string) (string, bool) {
	if strings.ContainsAny(name, `\/`) {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			abs, err := filepath.Abs(name)
			if err == nil {
				return abs, true
			}
			return name, true
		}
		return "", false
	}

	pathext := os.Getenv("PATHEXT")
	if pathext == "" {
		pathext = ".COM;.EXE;.BAT;.CMD"
	}
	exts := strings.Split(pathext, ";")

	hasExt := filepath.Ext(name) != ""

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		if hasExt {
			candidate := filepath.Join(dir, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, true
			}
			continue
		}
		for _, ext := range exts {
			candidate := filepath.Join(dir, name+ext)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, true
			}
		}
	}
	return "", false
}

func init() { command.Register(whichCommand{}) }
