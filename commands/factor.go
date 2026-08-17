package commands

import (
	"fmt"
	"strconv"

	"linuxcmd/internal/command"
)

type factorCommand struct{}

func (factorCommand) Name() string    { return "factor" }
func (factorCommand) Summary() string { return "print prime factors of an integer" }

func primeFactors(n int64) []int64 {
	var factors []int64
	for n%2 == 0 {
		factors = append(factors, 2)
		n /= 2
	}
	for d := int64(3); d*d <= n; d += 2 {
		for n%d == 0 {
			factors = append(factors, d)
			n /= d
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	return factors
}

func (factorCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: factor NUMBER...")
		return command.ExitUsage
	}
	exit := command.ExitSuccess
	for _, arg := range ctx.Args {
		n, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "factor: '%s' is not a positive integer\n", arg)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%d:", n)
		for _, f := range primeFactors(n) {
			fmt.Fprintf(ctx.Stdout, " %d", f)
		}
		fmt.Fprintln(ctx.Stdout)
	}
	return exit
}

func init() { command.Register(factorCommand{}) }
