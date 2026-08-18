package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
)

// This file holds the machinery shared by every linuxcmd command that
// deliberately does NOT reimplement its Linux counterpart, and instead
// hands off to a real, separately installed Windows program: ssh, make,
// python3, node, and so on. Those tools already ship first-class Windows
// builds, so shipping a half-working clone of them would be worse than
// useless. What linuxcmd adds is the *name*: "python3" works where only
// python.exe exists, "make" works where only mingw32-make.exe exists, and
// a missing tool produces a Linux-style diagnostic naming the installer
// instead of Windows' "'ssh' is not recognized as an internal or external
// command".
//
// The critical subtlety is self-recursion. The installer hardlinks the
// single linuxcmd binary to one .exe per registered command (see
// cmd/linuxcmd's doc comment), so after installing, LINUXCMD_HOME
// contains an ssh.exe that IS linuxcmd. If that directory precedes
// System32 on PATH -- which it usually does -- a naive PATH lookup for
// "ssh.exe" resolves straight back to this process and forks a loop.
// findExternal therefore rejects any candidate that is the same file as
// the running executable. Comparing by os.SameFile rather than by
// directory is what makes this correct: hardlinks share an identity but
// not a path, so a directory check alone (the approach git.go originally
// used) misses a hardlink placed anywhere else on PATH.

// externalCandidate is one Windows executable that can satisfy a Linux
// command name, plus any arguments that must precede the user's own.
type externalCandidate struct {
	// File is the executable's filename including extension, e.g.
	// "mingw32-make.exe" or "npm.cmd".
	File string
	// Args are prepended to the user's arguments. Used by the Python
	// launcher, where "py.exe" needs "-3" to mean python3.
	Args []string
}

// externalTool describes a linuxcmd command implemented as a handoff to
// an installed Windows program.
type externalTool struct {
	name    string
	summary string
	// candidates are tried in order; the first one found wins.
	candidates []externalCandidate
	// extraDirs are searched after PATH, for tools Windows ships outside
	// the default PATH (OpenSSH lives in System32\OpenSSH) or that
	// installers commonly leave off it.
	extraDirs func() []string
	// hint completes the sentence "<name>: not found on PATH (...)" and
	// should name the thing to install, not restate the problem.
	hint string
}

func (t externalTool) Name() string    { return t.name }
func (t externalTool) Summary() string { return t.summary }

func (t externalTool) Run(ctx *command.Context) int {
	exe, pre, ok := findExternal(t.candidates, t.extraDirs)
	if !ok {
		return externalNotFound(ctx, t.name, t.hint)
	}
	return runExternal(ctx, t.name, exe, append(append([]string{}, pre...), ctx.Args...))
}

// externalNotFound reports a missing external tool the way a Linux shell
// would, with the Windows-specific remedy appended. Windows' own message
// ("'ssh' is not recognized as an internal or external command") tells a
// user nothing about what to install, which is most of the value these
// wrappers add.
func externalNotFound(ctx *command.Context, prog, hint string) int {
	fmt.Fprintf(ctx.Stderr, "%s: command not found (%s)\n", prog, hint)
	return command.ExitNotFound
}

// findExternal locates the first available candidate, searching PATH and
// then extraDirs, skipping anything that is really linuxcmd itself.
// It returns the resolved path and the candidate's prefix arguments.
func findExternal(candidates []externalCandidate, extraDirs func() []string) (string, []string, bool) {
	dirs := filepath.SplitList(os.Getenv("PATH"))
	if extraDirs != nil {
		dirs = append(dirs, extraDirs()...)
	}

	var self os.FileInfo
	if exe, err := os.Executable(); err == nil {
		self, _ = os.Stat(exe)
	}

	for _, c := range candidates {
		for _, dir := range dirs {
			if dir == "" {
				continue
			}
			path := filepath.Join(dir, c.File)
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			// The candidate is one of linuxcmd's own hardlinks: keep
			// looking, or we would re-exec ourselves forever.
			if self != nil && os.SameFile(self, info) {
				continue
			}
			return path, c.Args, true
		}
	}
	return "", nil, false
}

// runExternal executes exe with args, wiring it to the caller's streams
// so interactive programs (ssh, vim, tmux) get the real console, and
// propagates the child's exit code as linuxcmd's own.
func runExternal(ctx *command.Context, prog, exe string, args []string) int {
	// CreateProcess cannot launch a .cmd or .bat directly -- npm and many
	// Node-ecosystem tools ship as .cmd shims -- so route those through
	// the command interpreter.
	var cmd *exec.Cmd
	switch strings.ToLower(filepath.Ext(exe)) {
	case ".cmd", ".bat":
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = filepath.Join(systemRoot(), "System32", "cmd.exe")
		}
		cmd = exec.Command(comspec, append([]string{"/c", exe}, args...)...)
	default:
		cmd = exec.Command(exe, args...)
	}
	cmd.Stdin = ctx.Stdin
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// exeCandidates is shorthand for the common case: plain executables with
// no prefix arguments, listed in preference order.
func exeCandidates(files ...string) []externalCandidate {
	out := make([]externalCandidate, 0, len(files))
	for _, f := range files {
		out = append(out, externalCandidate{File: f})
	}
	return out
}

// programFilesDirs returns the install roots joined with sub, covering
// both the 64-bit and 32-bit Program Files trees.
func programFilesDirs(sub ...string) []string {
	var dirs []string
	seen := map[string]bool{}
	for _, env := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
		root := os.Getenv(env)
		if root == "" || seen[strings.ToLower(root)] {
			continue
		}
		seen[strings.ToLower(root)] = true
		dirs = append(dirs, filepath.Join(append([]string{root}, sub...)...))
	}
	return dirs
}
