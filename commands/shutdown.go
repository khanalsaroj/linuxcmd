package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
)

// shutdownCommand wraps Windows' own shutdown.exe. Mutating the running
// machine's power state is inherently destructive, so this requires an
// explicit TIME argument (matching GNU shutdown's own requirement)
// rather than defaulting to an immediate action, and never runs unless
// the caller explicitly asked for it.
type shutdownCommand struct{}

func (shutdownCommand) Name() string    { return "shutdown" }
func (shutdownCommand) Summary() string { return "halt or reboot the machine via shutdown.exe" }

func shutdownExe() string {
	return filepath.Join(systemRoot(), "System32", "shutdown.exe")
}

// shutdownDelaySeconds maps GNU shutdown's TIME argument ("now", "+N" in
// minutes, or a bare number of seconds as a linuxcmd extension) onto the
// /t delay shutdown.exe expects. Absolute clock times ("HH:MM") aren't
// supported, since shutdown.exe only accepts a relative delay.
func shutdownDelaySeconds(t string) (int, error) {
	if t == "now" {
		return 0, nil
	}
	if strings.HasPrefix(t, "+") {
		minutes, err := strconv.Atoi(t[1:])
		if err != nil || minutes < 0 {
			return 0, fmt.Errorf("invalid time '%s'", t)
		}
		return minutes * 60, nil
	}
	seconds, err := strconv.Atoi(t)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("invalid time '%s' (use \"now\", \"+MINUTES\", or a bare second count)", t)
	}
	return seconds, nil
}

func (shutdownCommand) Run(ctx *command.Context) int {
	reboot := false
	var timeArg string
	for _, a := range ctx.Args {
		switch a {
		case "-r":
			reboot = true
		case "-h":
			reboot = false
		default:
			timeArg = a
		}
	}
	if timeArg == "" {
		fmt.Fprintln(ctx.Stderr, "usage: shutdown [-r|-h] TIME   (TIME: \"now\", \"+MINUTES\", or seconds)")
		return command.ExitUsage
	}
	delay, err := shutdownDelaySeconds(timeArg)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "shutdown: %s\n", err)
		return command.ExitUsage
	}

	winArgs := []string{"/t", strconv.Itoa(delay)}
	if reboot {
		winArgs = append([]string{"/r"}, winArgs...)
	} else {
		winArgs = append([]string{"/s"}, winArgs...)
	}

	cmd := exec.Command(shutdownExe(), winArgs...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "shutdown: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(shutdownCommand{}) }
