// Command linuxcmd is the single shared engine binary. Every per-command
// executable installed on PATH (ls.exe, cp.exe, ...) is a copy or
// hardlink of this exact binary; at startup it looks at the filename it
// was invoked as (os.Args[0]) to decide which Linux command to run. When
// invoked as "linuxcmd" itself, the command name is instead taken from
// the first argument, e.g. "linuxcmd ls -la".
//
// This "multicall binary" design (the same technique BusyBox uses) means
// there is exactly one compiled program regardless of how many commands
// are supported, keeping install size small and every command's startup
// fast.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
	_ "linuxcmd/commands" // registers every command via init()
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "linuxcmd: no arguments")
		return command.ExitUsage
	}

	invoked := invokedName(args[0])
	name := invoked
	rest := args[1:]

	if invoked == "linuxcmd" {
		if len(rest) == 0 {
			printUsage(os.Stdout)
			return command.ExitSuccess
		}
		switch rest[0] {
		case "-h", "--help":
			printUsage(os.Stdout)
			return command.ExitSuccess
		case "--list-commands":
			// Machine-readable command list (one name per line) consumed
			// by scripts/build.ps1 and installer/install.ps1, so both
			// discover commands from the registry instead of hardcoding
			// a duplicate list that would need updating by hand whenever
			// a command is added.
			for _, n := range command.Names() {
				fmt.Fprintln(os.Stdout, n)
			}
			return command.ExitSuccess
		}
		name = rest[0]
		rest = rest[1:]
	}

	cmd, ok := command.Lookup(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "linuxcmd: %s: command not found\n", name)
		return command.ExitNotFound
	}

	ctx := command.NewContext(name, rest)
	return cmd.Run(ctx)
}

// invokedName extracts the command name from an argv[0] like
// "C:\Program Files\LinuxCmd\ls.exe" -> "ls".
func invokedName(arg0 string) string {
	base := filepath.Base(arg0)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ToLower(base)
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, "linuxcmd - a Linux command compatibility layer for Windows CMD")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  linuxcmd <command> [arguments]")
	fmt.Fprintln(w, "  <command> [arguments]   (if <command>.exe is on PATH)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Available commands:")
	for _, n := range command.Names() {
		if c, ok := command.Lookup(n); ok {
			fmt.Fprintf(w, "  %-10s %s\n", n, c.Summary())
		}
	}
}
