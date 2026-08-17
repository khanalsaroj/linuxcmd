package commands

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"linuxcmd/internal/command"
)

type envCommand struct{}

func (envCommand) Name() string { return "env" }
func (envCommand) Summary() string {
	return "list environment variables or run a command with overrides"
}

func (envCommand) Run(ctx *command.Context) int {
	overrides := map[string]string{}
	i := 0
	for ; i < len(ctx.Args); i++ {
		eq := strings.IndexByte(ctx.Args[i], '=')
		if eq < 0 {
			break
		}
		overrides[ctx.Args[i][:eq]] = ctx.Args[i][eq+1:]
	}

	if i >= len(ctx.Args) {
		env := os.Environ()
		for k, v := range overrides {
			env = append(env, k+"="+v)
		}
		sort.Strings(env)
		for _, kv := range env {
			fmt.Fprintln(ctx.Stdout, kv)
		}
		return command.ExitSuccess
	}

	cmd := exec.Command(ctx.Args[i], ctx.Args[i+1:]...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	cmd.Env = os.Environ()
	for k, v := range overrides {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "env: %s\n", err)
		return command.ExitNotFound
	}
	return command.ExitSuccess
}

func init() { command.Register(envCommand{}) }
