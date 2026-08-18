package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/paths"
)

// xdg-open hands a file or URL to whatever application the desktop has
// registered for it. Windows' equivalent is ShellExecute, which is the
// same mechanism Explorer uses for a double-click, so this is a genuine
// one-to-one mapping rather than an approximation. The command is also
// registered as "open" because that is the name macOS users reach for
// and there is no Windows program by that name to shadow.
type xdgOpenCommand struct {
	name string
}

func (c xdgOpenCommand) Name() string { return c.name }
func (c xdgOpenCommand) Summary() string {
	return "open a file or URL in the default application"
}

// xdg-open's documented exit codes. Scripts branch on these, so they are
// reproduced rather than collapsed into a generic failure.
const (
	xdgExitSyntax     = 1 // error in command line syntax
	xdgExitNotFound   = 2 // file does not exist
	xdgExitNoHandler  = 3 // no application available to open the file
	xdgExitOpenFailed = 4 // the action failed
)

var (
	shell32          = syscall.NewLazyDLL("shell32.dll")
	procShellExecute = shell32.NewProc("ShellExecuteW")
)

// ShellExecute return values below 33 are error codes rather than
// instance handles.
const (
	seErrFileNotFound = 2
	seErrPathNotFound = 3
	seErrAccessDenied = 5
	seErrNoAssoc      = 31
	swShowNormal      = 1
)

func (c xdgOpenCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintf(ctx.Stderr, "usage: %s FILE|URL\n", c.name)
		return xdgExitSyntax
	}

	exit := command.ExitSuccess
	for _, arg := range ctx.Args {
		if code := c.openOne(ctx, arg); code != command.ExitSuccess {
			exit = code
		}
	}
	return exit
}

func (c xdgOpenCommand) openOne(ctx *command.Context, arg string) int {
	target := arg

	// URLs and other schemes (mailto:, ms-settings:) must reach
	// ShellExecute untouched; only real paths get Linux-to-Windows
	// translation, and only those are checked for existence.
	if !hasURIScheme(arg) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s: cannot resolve path\n", c.name, arg)
			return xdgExitNotFound
		}
		if _, err := os.Stat(resolved); err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s: No such file or directory\n", c.name, arg)
			return xdgExitNotFound
		}
		target = resolved
	}

	ret, err := shellExecute(target)
	if ret > 32 {
		return command.ExitSuccess
	}

	switch ret {
	case seErrNoAssoc:
		fmt.Fprintf(ctx.Stderr, "%s: %s: no application is associated with this file type\n", c.name, arg)
		return xdgExitNoHandler
	case seErrFileNotFound, seErrPathNotFound:
		fmt.Fprintf(ctx.Stderr, "%s: %s: No such file or directory\n", c.name, arg)
		return xdgExitNotFound
	case seErrAccessDenied:
		fmt.Fprintf(ctx.Stderr, "%s: %s: Permission denied\n", c.name, arg)
		return xdgExitOpenFailed
	default:
		fmt.Fprintf(ctx.Stderr, "%s: %s: %s\n", c.name, arg, err)
		return xdgExitOpenFailed
	}
}

func shellExecute(target string) (uintptr, error) {
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return 0, err
	}
	file, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return 0, err
	}
	ret, _, callErr := procShellExecute.Call(
		0, // no parent window
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0, // no parameters
		0, // inherit the working directory
		swShowNormal,
	)
	return ret, callErr
}

// hasURIScheme reports whether s looks like "scheme:..." rather than a
// filesystem path. A single letter before the colon is excluded so that
// Windows drive paths such as "C:\Users" are not mistaken for schemes.
func hasURIScheme(s string) bool {
	colon := strings.IndexByte(s, ':')
	if colon < 2 {
		return false
	}
	for i := 0; i < colon; i++ {
		c := s[i]
		isSchemeChar := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '-' || c == '.'
		if !isSchemeChar {
			return false
		}
	}
	return true
}

func init() {
	command.Register(xdgOpenCommand{name: "xdg-open"})
	command.Register(xdgOpenCommand{name: "open"})
}
