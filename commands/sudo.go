package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// sudo and su raise privileges. Both work on Windows, but with one
// difference large enough that it is printed as a warning rather than
// buried in documentation: an elevated process cannot inherit the
// current console's handles, because elevation crosses a security
// boundary. Windows starts it in a NEW console instead. That means
// redirection and pipes do not reach across the elevation:
//
//	sudo ls > out.txt        # out.txt stays empty
//	sudo cat file | grep x   # grep sees nothing
//
// Windows 11 ships its own sudo.exe, which solves this properly with an
// inline mode. When it is present, these commands hand off to it so the
// behavior matches the system tool exactly, and the caveat above does
// not apply.

type sudoCommand struct{}

func (sudoCommand) Name() string    { return "sudo" }
func (sudoCommand) Summary() string { return "run a command with administrator rights" }

var sudoSpec = parser.Spec{
	{Short: 'n', Long: "non-interactive"},
	{Short: 'v', Long: "validate"},
	{Short: 'k', Long: "reset-timestamp"},
}

const (
	seVerbRunAs      = "runas"
	shellExecuteOK   = 32 // ShellExecute returns >32 on success
	seErrCancelledOp = 1223
)

func (sudoCommand) Run(ctx *command.Context) int {
	// Options are recognized so that scripts using them do not fail
	// outright, but sudo's credential caching has no Windows analogue:
	// UAC decides per launch.
	res, err := parser.Parse(ctx.Args, sudoSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "sudo: %s\n", err)
		return command.ExitUsage
	}
	if res.Bool('v', "validate") || res.Bool('k', "reset-timestamp") {
		fmt.Fprintln(ctx.Stderr, "sudo: there is no credential cache on Windows; UAC prompts per launch")
		return command.ExitSuccess
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: sudo COMMAND [ARGUMENT]...")
		return command.ExitUsage
	}

	// Prefer the system sudo when this Windows has one: it can run the
	// elevated command inline, which nothing else here can do.
	if exe, _, ok := findExternal(exeCandidates("sudo.exe"), systemSudoDirs); ok {
		return runExternal(ctx, "sudo", exe, ctx.Args)
	}

	target, args, err := resolveElevationTarget(res.Positional)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "sudo: %s\n", err)
		return command.ExitNotFound
	}

	if res.Bool('n', "non-interactive") {
		// -n means "fail rather than prompt", and elevation here always
		// prompts.
		fmt.Fprintln(ctx.Stderr, "sudo: a password is required (UAC cannot be suppressed)")
		return command.ExitFailure
	}

	fmt.Fprintln(ctx.Stderr, "sudo: elevating in a new console; output will not return to this one")
	return elevate(ctx, "sudo", target, args)
}

// systemSudoDirs points at the native Windows sudo, which lives in
// System32 and may not be on PATH.
func systemSudoDirs() []string {
	return []string{filepath.Join(systemRoot(), "System32")}
}

// resolveElevationTarget works out which executable to elevate. A
// linuxcmd command name resolves to this binary invoked with that name,
// so "sudo rm -rf dir" elevates linuxcmd's own rm rather than failing to
// find an rm.exe.
func resolveElevationTarget(argv []string) (string, []string, error) {
	name := argv[0]
	rest := argv[1:]

	if path, err := exec.LookPath(name); err == nil {
		return path, rest, nil
	}
	if _, registered := command.Lookup(name); registered {
		self, err := os.Executable()
		if err != nil {
			return "", nil, fmt.Errorf("cannot locate linuxcmd itself: %w", err)
		}
		return self, append([]string{name}, rest...), nil
	}
	return "", nil, fmt.Errorf("%s: command not found", name)
}

// elevate launches target through the UAC consent prompt.
func elevate(ctx *command.Context, prog, target string, args []string) int {
	verb, err := syscall.UTF16PtrFromString(seVerbRunAs)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}
	file, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}

	var paramsPtr *uint16
	if len(args) > 0 {
		paramsPtr, err = syscall.UTF16PtrFromString(quoteCommandLine(args))
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
			return command.ExitFailure
		}
	}

	// The elevated process starts in the current directory so that
	// relative path arguments still mean what the caller meant.
	var dirPtr *uint16
	if cwd, err := os.Getwd(); err == nil {
		dirPtr, _ = syscall.UTF16PtrFromString(cwd)
	}

	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(paramsPtr)),
		uintptr(unsafe.Pointer(dirPtr)),
		swShowNormal,
	)
	if ret > shellExecuteOK {
		return command.ExitSuccess
	}
	if ret == seErrCancelledOp {
		fmt.Fprintf(ctx.Stderr, "%s: elevation was declined\n", prog)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stderr, "%s: cannot elevate (error %d)\n", prog, ret)
	return command.ExitFailure
}

// quoteCommandLine rebuilds a Windows command-line string from already
// split arguments, quoting the ones that need it. ShellExecute takes
// parameters as a single string, so the splitting Windows did on the way
// in has to be undone on the way out.
func quoteCommandLine(args []string) string {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if a == "" {
			parts = append(parts, `""`)
			continue
		}
		if strings.ContainsAny(a, " \t\"") {
			parts = append(parts, `"`+strings.ReplaceAll(a, `"`, `\"`)+`"`)
			continue
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}

// --- su ------------------------------------------------------------------

type suCommand struct{}

func (suCommand) Name() string    { return "su" }
func (suCommand) Summary() string { return "start a shell as another user" }

var suSpec = parser.Spec{
	{Short: 'l', Long: "login"},
	{Short: 'c', HasArg: true},
}

func (suCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, suSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "su: %s\n", err)
		return command.ExitUsage
	}

	// Bare "su" means root on Linux; the Windows equivalent is
	// elevation, not switching to a differently named account.
	if len(res.Positional) == 0 {
		shell := os.Getenv("COMSPEC")
		if shell == "" {
			shell = filepath.Join(systemRoot(), "System32", "cmd.exe")
		}
		var args []string
		if cmdline, ok := res.Value('c', ""); ok {
			args = []string{"/c", cmdline}
		}
		fmt.Fprintln(ctx.Stderr, "su: elevating in a new console; Windows has no root account to switch to")
		return elevate(ctx, "su", shell, args)
	}

	user := res.Positional[0]
	runas := filepath.Join(systemRoot(), "System32", "runas.exe")
	if _, err := os.Stat(runas); err != nil {
		fmt.Fprintln(ctx.Stderr, "su: runas.exe is not available on this system")
		return command.ExitNotFound
	}

	shell := os.Getenv("COMSPEC")
	if shell == "" {
		shell = filepath.Join(systemRoot(), "System32", "cmd.exe")
	}
	target := shell
	if cmdline, ok := res.Value('c', ""); ok {
		target = shell + " /c " + cmdline
	}

	// runas always collects the password through its own interactive
	// prompt. linuxcmd never reads or forwards a password itself.
	fmt.Fprintln(ctx.Stderr, "su: runas will prompt for the account password")
	return runExternal(ctx, "su", runas, []string{"/user:" + user, target})
}

func init() {
	command.Register(sudoCommand{})
	command.Register(suCommand{})
}
