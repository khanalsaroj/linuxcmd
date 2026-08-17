package commands

import (
	"fmt"
	"strconv"

	"linuxcmd/internal/command"
)

type seqCommand struct{}

func (seqCommand) Name() string    { return "seq" }
func (seqCommand) Summary() string { return "print a numeric sequence" }

func (seqCommand) Run(ctx *command.Context) int {
	var start, step, end float64 = 1, 1, 0
	var err error

	switch len(ctx.Args) {
	case 1:
		end, err = strconv.ParseFloat(ctx.Args[0], 64)
	case 2:
		start, err = strconv.ParseFloat(ctx.Args[0], 64)
		if err == nil {
			end, err = strconv.ParseFloat(ctx.Args[1], 64)
		}
	case 3:
		start, err = strconv.ParseFloat(ctx.Args[0], 64)
		if err == nil {
			step, err = strconv.ParseFloat(ctx.Args[1], 64)
		}
		if err == nil {
			end, err = strconv.ParseFloat(ctx.Args[2], 64)
		}
	default:
		fmt.Fprintln(ctx.Stderr, "usage: seq [FIRST [STEP]] LAST")
		return command.ExitUsage
	}
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "seq: invalid number\n")
		return command.ExitUsage
	}
	if step == 0 {
		fmt.Fprintln(ctx.Stderr, "seq: step must not be zero")
		return command.ExitUsage
	}

	if step > 0 {
		for v := start; v <= end; v += step {
			printSeqValue(ctx, v)
		}
	} else {
		for v := start; v >= end; v += step {
			printSeqValue(ctx, v)
		}
	}
	return command.ExitSuccess
}

func printSeqValue(ctx *command.Context, v float64) {
	if v == float64(int64(v)) {
		fmt.Fprintf(ctx.Stdout, "%d\n", int64(v))
	} else {
		fmt.Fprintf(ctx.Stdout, "%g\n", v)
	}
}

func init() { command.Register(seqCommand{}) }
