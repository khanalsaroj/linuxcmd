package commands

import (
	"fmt"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/paths"
)

type testCommand struct{}

func (testCommand) Name() string    { return "test" }
func (testCommand) Summary() string { return "evaluate a conditional expression" }

func (testCommand) Run(ctx *command.Context) int {
	args := ctx.Args
	// "[" is a common alias for test whose closing "]" argument should
	// be stripped before evaluation, in case a shell alias forwards it.
	if len(args) > 0 && args[len(args)-1] == "]" {
		args = args[:len(args)-1]
	}

	negate := false
	if len(args) > 0 && args[0] == "!" {
		negate = true
		args = args[1:]
	}

	result, err := evalTest(args)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "test: %s\n", err)
		return command.ExitUsage
	}
	if negate {
		result = !result
	}
	if result {
		return command.ExitSuccess
	}
	return command.ExitFailure
}

func evalTest(args []string) (bool, error) {
	switch len(args) {
	case 0:
		return false, nil
	case 1:
		return args[0] != "", nil
	case 2:
		return evalUnary(args[0], args[1])
	case 3:
		return evalBinary(args[0], args[1], args[2])
	default:
		return false, fmt.Errorf("too many arguments")
	}
}

func evalUnary(op, arg string) (bool, error) {
	switch op {
	case "-z":
		return arg == "", nil
	case "-n":
		return arg != "", nil
	case "-e":
		return statExists(arg), nil
	case "-f":
		return statIs(arg, false), nil
	case "-d":
		return statIs(arg, true), nil
	case "-r", "-w", "-x":
		return statExists(arg), nil
	case "-s":
		info, err := os.Stat(resolveOrRaw(arg))
		return err == nil && info.Size() > 0, nil
	default:
		return false, fmt.Errorf("unknown unary operator '%s'", op)
	}
}

func evalBinary(a, op, b string) (bool, error) {
	switch op {
	case "=", "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case "-eq", "-ne", "-lt", "-le", "-gt", "-ge":
		na, err1 := strconv.ParseInt(a, 10, 64)
		nb, err2 := strconv.ParseInt(b, 10, 64)
		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("integer expression expected")
		}
		switch op {
		case "-eq":
			return na == nb, nil
		case "-ne":
			return na != nb, nil
		case "-lt":
			return na < nb, nil
		case "-le":
			return na <= nb, nil
		case "-gt":
			return na > nb, nil
		case "-ge":
			return na >= nb, nil
		}
	}
	return false, fmt.Errorf("unknown operator '%s'", op)
}

func resolveOrRaw(p string) string {
	resolved, err := paths.Resolve(p)
	if err != nil {
		return p
	}
	return resolved
}

func statExists(p string) bool {
	_, err := os.Stat(resolveOrRaw(p))
	return err == nil
}

func statIs(p string, dir bool) bool {
	info, err := os.Stat(resolveOrRaw(p))
	if err != nil {
		return false
	}
	return info.IsDir() == dir
}

func init() { command.Register(testCommand{}) }
