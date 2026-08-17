package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"linuxcmd/internal/command"
)

// routeCommand wraps Windows' own route.exe. Windows route.exe uses a
// verb-first syntax ("route print", "route add ...") rather than Linux
// route's flag-first syntax, so a bare "route" or "route -n" maps to
// "route print"; anything else (add/delete/change) passes through
// untranslated -- route.exe itself already requires administrator rights
// for mutations, so no extra gating is added here.
type routeCommand struct{}

func (routeCommand) Name() string    { return "route" }
func (routeCommand) Summary() string { return "show or modify the IP routing table" }

func (routeCommand) Run(ctx *command.Context) int {
	winArgs := []string{"print"}
	if len(ctx.Args) > 0 && ctx.Args[0] != "-n" {
		winArgs = ctx.Args
	}
	routeExe := filepath.Join(systemRoot(), "System32", "route.exe")
	cmd := exec.Command(routeExe, winArgs...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "route: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(routeCommand{}) }
