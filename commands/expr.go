package commands

import (
	"fmt"
	"regexp"
	"strconv"

	"linuxcmd/internal/command"
)

type exprCommand struct{}

func (exprCommand) Name() string    { return "expr" }
func (exprCommand) Summary() string { return "evaluate a simple integer, string, or regex expression" }

func (exprCommand) Run(ctx *command.Context) int {
	args := ctx.Args
	if len(args) == 1 {
		return printExprResult(ctx, args[0])
	}
	if len(args) != 3 {
		fmt.Fprintln(ctx.Stderr, "usage: expr EXPRESSION")
		return command.ExitUsage
	}

	a, op, b := args[0], args[1], args[2]

	switch op {
	case "+", "-", "*", "/", "%":
		na, err1 := strconv.Atoi(a)
		nb, err2 := strconv.Atoi(b)
		if err1 != nil || err2 != nil {
			fmt.Fprintln(ctx.Stderr, "expr: non-integer argument")
			return command.ExitUsage
		}
		var result int
		switch op {
		case "+":
			result = na + nb
		case "-":
			result = na - nb
		case "*":
			result = na * nb
		case "/", "%":
			if nb == 0 {
				fmt.Fprintln(ctx.Stderr, "expr: division by zero")
				return command.ExitFailure
			}
			if op == "/" {
				result = na / nb
			} else {
				result = na % nb
			}
		}
		return printExprResult(ctx, strconv.Itoa(result))

	case "=", "==":
		return printExprBool(ctx, a == b)
	case "!=":
		return printExprBool(ctx, a != b)
	case "<", "<=", ">", ">=":
		return printExprCompare(ctx, a, b, op)
	case ":":
		re, err := regexp.Compile("^" + b)
		if err != nil {
			fmt.Fprintln(ctx.Stderr, "expr: invalid regex")
			return command.ExitUsage
		}
		match := re.FindString(a)
		return printExprResult(ctx, strconv.Itoa(len(match)))
	default:
		fmt.Fprintf(ctx.Stderr, "expr: unknown operator '%s'\n", op)
		return command.ExitUsage
	}
}

func printExprResult(ctx *command.Context, s string) int {
	fmt.Fprintln(ctx.Stdout, s)
	if s == "" || s == "0" {
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func printExprBool(ctx *command.Context, b bool) int {
	if b {
		return printExprResult(ctx, "1")
	}
	return printExprResult(ctx, "0")
}

func printExprCompare(ctx *command.Context, a, b, op string) int {
	na, err1 := strconv.Atoi(a)
	nb, err2 := strconv.Atoi(b)
	var result bool
	if err1 == nil && err2 == nil {
		switch op {
		case "<":
			result = na < nb
		case "<=":
			result = na <= nb
		case ">":
			result = na > nb
		case ">=":
			result = na >= nb
		}
	} else {
		switch op {
		case "<":
			result = a < b
		case "<=":
			result = a <= b
		case ">":
			result = a > b
		case ">=":
			result = a >= b
		}
	}
	return printExprBool(ctx, result)
}

func init() { command.Register(exprCommand{}) }
