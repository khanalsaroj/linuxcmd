package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
)

// gitCommand is a transparent wrapper around the real, separately
// installed git.exe -- reimplementing Git is out of scope. Because the
// installer creates a git.exe alongside every other registered command
// (see cmd/linuxcmd's doc comment), a plain PATH lookup for "git" could
// resolve back to this very binary if linuxcmd's own install directory
// happens to be searched first. findRealGit explicitly skips linuxcmd's
// own directory to avoid that recursion, the same problem ping.go solves
// by hardcoding ping.exe's System32 path.
type gitCommand struct{}

func (gitCommand) Name() string    { return "git" }
func (gitCommand) Summary() string { return "wrap the installed git.exe" }

func (gitCommand) Run(ctx *command.Context) int {
	gitExe, ok := findRealGit()
	if !ok {
		fmt.Fprintln(ctx.Stderr, "git: no git.exe found on PATH (install Git for Windows)")
		return command.ExitNotFound
	}
	cmd := exec.Command(gitExe, ctx.Args...)
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr
	cmd.Stdin = ctx.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "git: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func findRealGit() (string, bool) {
	selfDir := ""
	if exe, err := os.Executable(); err == nil {
		selfDir = filepath.Dir(exe)
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" || strings.EqualFold(dir, selfDir) {
			continue
		}
		candidate := filepath.Join(dir, "git.exe")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

func init() { command.Register(gitCommand{}) }
