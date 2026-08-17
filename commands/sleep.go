package commands

import (
	"fmt"
	"strconv"
	"time"

	"linuxcmd/internal/command"
)

type sleepCommand struct{}

func (sleepCommand) Name() string    { return "sleep" }
func (sleepCommand) Summary() string { return "pause for a duration" }

func (sleepCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: sleep NUMBER[smhd]...")
		return command.ExitUsage
	}

	var total time.Duration
	for _, arg := range ctx.Args {
		d, err := parseSleepDuration(arg)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "sleep: invalid time interval '%s'\n", arg)
			return command.ExitUsage
		}
		total += d
	}

	time.Sleep(total)
	return command.ExitSuccess
}

// parseSleepDuration accepts a plain number (seconds, fractional allowed)
// or a number with a single Linux-style suffix: s(econds), m(inutes),
// h(ours), d(ays).
func parseSleepDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty interval")
	}
	unit := time.Second
	numPart := s
	switch s[len(s)-1] {
	case 's':
		unit = time.Second
		numPart = s[:len(s)-1]
	case 'm':
		unit = time.Minute
		numPart = s[:len(s)-1]
	case 'h':
		unit = time.Hour
		numPart = s[:len(s)-1]
	case 'd':
		unit = 24 * time.Hour
		numPart = s[:len(s)-1]
	}
	n, err := strconv.ParseFloat(numPart, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid interval")
	}
	return time.Duration(n * float64(unit)), nil
}

func init() { command.Register(sleepCommand{}) }
