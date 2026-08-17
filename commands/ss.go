package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// ssCommand has no direct Windows equivalent, so it wraps netstat.exe's
// output and filters it client-side to approximate ss's common flags.
type ssCommand struct{}

func (ssCommand) Name() string    { return "ss" }
func (ssCommand) Summary() string { return "show sockets (netstat-based approximation)" }

var ssSpec = parser.Spec{
	{Short: 't'},
	{Short: 'u'},
	{Short: 'l'},
	{Short: 'n'},
}

func (ssCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, ssSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ss: %s\n", err)
		return command.ExitUsage
	}
	tcp := res.Bool('t', "")
	udp := res.Bool('u', "")
	listenOnly := res.Bool('l', "")

	netstatExe := filepath.Join(systemRoot(), "System32", "netstat.exe")
	cmd := exec.Command(netstatExe, "-a", "-n")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = ctx.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(ctx.Stderr, "ss: %s\n", err)
		return command.ExitFailure
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		upper := strings.ToUpper(line)
		if tcp && !strings.Contains(upper, "TCP") {
			continue
		}
		if udp && !strings.Contains(upper, "UDP") {
			continue
		}
		if listenOnly && !strings.Contains(upper, "LISTENING") && !strings.Contains(upper, "UDP") {
			continue
		}
		fmt.Fprintln(ctx.Stdout, line)
	}
	return command.ExitSuccess
}

func init() { command.Register(ssCommand{}) }
