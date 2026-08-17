package commands

import (
	"fmt"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type pgrepCommand struct{}

func (pgrepCommand) Name() string    { return "pgrep" }
func (pgrepCommand) Summary() string { return "find process IDs by name" }

var pgrepSpec = parser.Spec{
	{Short: 'f'},
	{Short: 'l', Long: "list-name"},
}

func (pgrepCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, pgrepSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pgrep: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: pgrep [-l] PATTERN")
		return command.ExitUsage
	}
	pattern := strings.ToLower(res.Positional[0])
	showName := res.Bool('l', "list-name")

	procs, err := snapshotProcesses()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pgrep: %s\n", err)
		return command.ExitFailure
	}

	matched := false
	for _, p := range procs {
		if !strings.Contains(strings.ToLower(p.Name), pattern) {
			continue
		}
		matched = true
		if showName {
			fmt.Fprintf(ctx.Stdout, "%d %s\n", p.PID, p.Name)
		} else {
			fmt.Fprintln(ctx.Stdout, strconv.FormatUint(uint64(p.PID), 10))
		}
	}
	if !matched {
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(pgrepCommand{}) }
