package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
)

// pingCommand wraps Windows' own ping.exe rather than reimplementing
// ICMP: raw ICMP sockets need elevated privileges on Windows, and the
// spec explicitly allows os/exec (with separated arguments, never a
// shell string) for commands that genuinely need another executable.
//
// It resolves ping.exe by an explicit System32 path instead of a PATH
// lookup: since our own install directory (which also contains a
// ping.exe) is on PATH, an unqualified exec.Command("ping", ...) could
// resolve back to *this* binary and recurse into itself.
type pingCommand struct{}

func (pingCommand) Name() string    { return "ping" }
func (pingCommand) Summary() string { return "send ICMP echo requests to a host" }

var linuxToWindowsPingFlags = map[string]string{
	"-c": "-n", // Linux: packet count -> Windows: packet count
}

func (pingCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: ping [-c count] host")
		return command.ExitUsage
	}

	var winArgs []string
	host := ""
	for i := 0; i < len(ctx.Args); i++ {
		a := ctx.Args[i]
		if mapped, ok := linuxToWindowsPingFlags[a]; ok && i+1 < len(ctx.Args) {
			winArgs = append(winArgs, mapped, ctx.Args[i+1])
			i++
			continue
		}
		if !strings.HasPrefix(a, "-") {
			host = a
			continue
		}
		winArgs = append(winArgs, a) // pass unrecognized flags through; ping.exe validates them
	}
	if host == "" {
		fmt.Fprintln(ctx.Stderr, "ping: usage error: destination address required")
		return command.ExitUsage
	}
	winArgs = append(winArgs, host)

	pingExe := filepath.Join(systemRoot(), "System32", "PING.EXE")
	cmd := exec.Command(pingExe, winArgs...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "ping: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func systemRoot() string {
	if v := os.Getenv("SystemRoot"); v != "" {
		return v
	}
	return `C:\Windows`
}

func init() { command.Register(pingCommand{}) }
