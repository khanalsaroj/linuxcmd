package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// systemctlCommand maps a small subset of subcommands onto sc.exe, the
// native Windows service control tool. start/stop/restart require
// administrator rights on Windows just as changing a systemd unit's
// state does on Linux; sc.exe itself enforces that, so no separate check
// is added here.
type systemctlCommand struct{}

func (systemctlCommand) Name() string { return "systemctl" }
func (systemctlCommand) Summary() string {
	return "control Windows services (status/start/stop/restart)"
}

func scExe() string {
	return filepath.Join(systemRoot(), "System32", "sc.exe")
}

func runSC(ctx *command.Context, args ...string) int {
	cmd := exec.Command(scExe(), args...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "%s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func (systemctlCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: systemctl {status|start|stop|restart} SERVICE")
		return command.ExitUsage
	}
	verb, name := ctx.Args[0], ctx.Args[1]

	switch verb {
	case "status":
		return runSC(ctx, "query", name)
	case "start":
		return runSC(ctx, "start", name)
	case "stop":
		return runSC(ctx, "stop", name)
	case "restart":
		runSC(ctx, "stop", name)
		return runSC(ctx, "start", name)
	default:
		fmt.Fprintf(ctx.Stderr, "systemctl: unsupported verb '%s'\n", verb)
		return command.ExitUsage
	}
}

func init() { command.Register(systemctlCommand{}) }
