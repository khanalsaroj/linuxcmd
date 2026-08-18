package commands

import (
	"linuxcmd/internal/command"
)

// gitCommand is a transparent wrapper around the real, separately
// installed git.exe -- reimplementing Git is out of scope. The lookup,
// including the self-recursion guard that keeps linuxcmd's own git.exe
// hardlink from re-exec'ing this process, lives in passthrough.go and is
// shared with every other handoff command (ssh, make, node, ...).
type gitCommand struct{}

func (gitCommand) Name() string    { return "git" }
func (gitCommand) Summary() string { return "wrap the installed git.exe" }

func (gitCommand) Run(ctx *command.Context) int {
	gitExe, ok := findRealGit()
	if !ok {
		return externalNotFound(ctx, "git", "install Git for Windows")
	}
	return runExternal(ctx, "git", gitExe, ctx.Args)
}

func findRealGit() (string, bool) {
	exe, _, ok := findExternal(exeCandidates("git.exe"), gitExtraDirs)
	return exe, ok
}

func gitExtraDirs() []string { return programFilesDirs("Git", "cmd") }

func init() { command.Register(gitCommand{}) }
