package commands

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// topCommand is a refreshing process monitor. Real "top" refreshes
// forever until interrupted; this supports the same -n/-d flags GNU top
// does, so a script (or a test) can request a bounded number of
// refreshes instead of relying on Ctrl+C.
type topCommand struct{}

func (topCommand) Name() string    { return "top" }
func (topCommand) Summary() string { return "display a refreshing process list" }

var topSpec = parser.Spec{
	{Short: 'n', HasArg: true},
	{Short: 'd', HasArg: true},
}

func (topCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, topSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "top: %s\n", err)
		return command.ExitUsage
	}

	iterations := -1 // -1 means refresh forever, like real top
	if v, ok := res.Value('n', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "top: invalid iteration count '%s'\n", v)
			return command.ExitUsage
		}
		iterations = n
	}
	delay := 2 * time.Second
	if v, ok := res.Value('d', ""); ok {
		secs, err := strconv.ParseFloat(v, 64)
		if err != nil || secs <= 0 {
			fmt.Fprintf(ctx.Stderr, "top: invalid delay '%s'\n", v)
			return command.ExitUsage
		}
		delay = time.Duration(secs * float64(time.Second))
	}

	for i := 0; iterations < 0 || i < iterations; i++ {
		procs, err := snapshotProcesses()
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "top: %s\n", err)
			return command.ExitFailure
		}
		sort.Slice(procs, func(a, b int) bool { return procs[a].PID < procs[b].PID })

		fmt.Fprintf(ctx.Stdout, "%8s %8s %s\n", "PID", "PPID", "CMD")
		for _, p := range procs {
			fmt.Fprintf(ctx.Stdout, "%8d %8d %s\n", p.PID, p.PPID, p.Name)
		}

		if iterations < 0 || i < iterations-1 {
			time.Sleep(delay)
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(topCommand{}) }
