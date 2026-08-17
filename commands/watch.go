package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// watchCommand re-runs a command periodically. Real "watch" refreshes
// forever until interrupted; -c (a linuxcmd extension, since GNU watch
// has no such flag) bounds the iteration count for scripted/test use,
// the same way top's -n does.
type watchCommand struct{}

func (watchCommand) Name() string    { return "watch" }
func (watchCommand) Summary() string { return "re-run a command periodically" }

var watchSpec = parser.Spec{
	{Short: 'n', HasArg: true},
	{Short: 'c', HasArg: true},
}

func (watchCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, watchSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "watch: %s\n", err)
		return command.ExitUsage
	}
	interval := 2 * time.Second
	if v, ok := res.Value('n', ""); ok {
		secs, err := strconv.ParseFloat(v, 64)
		if err != nil || secs <= 0 {
			fmt.Fprintf(ctx.Stderr, "watch: invalid interval '%s'\n", v)
			return command.ExitUsage
		}
		interval = time.Duration(secs * float64(time.Second))
	}
	iterations := -1
	if v, ok := res.Value('c', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "watch: invalid count '%s'\n", v)
			return command.ExitUsage
		}
		iterations = n
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: watch [-n SECONDS] [-c COUNT] COMMAND [ARG...]")
		return command.ExitUsage
	}

	for i := 0; iterations < 0 || i < iterations; i++ {
		enableVirtualTerminal()
		fmt.Fprint(ctx.Stdout, "\x1b[2J\x1b[H")

		cmd := exec.Command(res.Positional[0], res.Positional[1:]...)
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
		cmd.Run()

		if iterations < 0 || i < iterations-1 {
			time.Sleep(interval)
		}
	}
	return command.ExitSuccess
}

func init() { command.Register(watchCommand{}) }
