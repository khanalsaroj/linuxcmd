package commands

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"

	"linuxcmd/internal/command"
)

type xargsCommand struct{}

func (xargsCommand) Name() string    { return "xargs" }
func (xargsCommand) Summary() string { return "build and run commands from stdin" }

func (xargsCommand) Run(ctx *command.Context) int {
	// xargs' own command line mixes its flags with the target command's
	// (whose flags must pass through untouched), so this only recognizes
	// a leading "-n COUNT" rather than using the shared parser.Spec.
	args := ctx.Args
	batchSize := -1
	i := 0
	if i < len(args) && args[i] == "-n" && i+1 < len(args) {
		n, err := strconv.Atoi(args[i+1])
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "xargs: invalid -n value '%s'\n", args[i+1])
			return command.ExitUsage
		}
		batchSize = n
		i += 2
	}

	prog := "echo"
	var baseArgs []string
	if i < len(args) {
		prog = args[i]
		baseArgs = args[i+1:]
	}

	scanner := bufio.NewScanner(ctx.Stdin)
	scanner.Split(bufio.ScanWords)
	var tokens []string
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}
	if len(tokens) == 0 {
		return command.ExitSuccess
	}

	run := func(extra []string) int {
		cmd := exec.Command(prog, append(append([]string{}, baseArgs...), extra...)...)
		cmd.Stdout = ctx.Stdout
		cmd.Stderr = ctx.Stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode()
			}
			fmt.Fprintf(ctx.Stderr, "xargs: %s\n", err)
			return command.ExitNotFound
		}
		return command.ExitSuccess
	}

	if batchSize <= 0 {
		return run(tokens)
	}
	exit := command.ExitSuccess
	for start := 0; start < len(tokens); start += batchSize {
		end := start + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		if code := run(tokens[start:end]); code != 0 {
			exit = code
		}
	}
	return exit
}

func init() { command.Register(xargsCommand{}) }
