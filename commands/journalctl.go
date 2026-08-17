package commands

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// journalctlCommand wraps wevtutil.exe against the Application log (the
// closest analog of the general-purpose systemd journal), supporting
// only "-n COUNT" the way "journalctl -n 50" is most often used.
type journalctlCommand struct{}

func (journalctlCommand) Name() string    { return "journalctl" }
func (journalctlCommand) Summary() string { return "show recent Application log entries" }

var journalctlSpec = parser.Spec{
	{Short: 'n', HasArg: true},
}

func (journalctlCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, journalctlSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "journalctl: %s\n", err)
		return command.ExitUsage
	}
	count := 50
	if v, ok := res.Value('n', ""); ok {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fmt.Fprintf(ctx.Stderr, "journalctl: invalid count '%s'\n", v)
			return command.ExitUsage
		}
		count = n
	}

	wevtutil := filepath.Join(systemRoot(), "System32", "wevtutil.exe")
	cmd := exec.Command(wevtutil, "qe", "Application", "/c:"+strconv.Itoa(count), "/rd:true", "/f:text")
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "journalctl: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(journalctlCommand{}) }
