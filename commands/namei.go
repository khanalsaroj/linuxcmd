package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// nameiCommand walks a pathname one component at a time. On Linux that
// is a niche debugging tool; here it is the most direct way to see what
// linuxcmd actually did with a path, because every argument goes through
// the same Linux-to-Windows translation (~, /tmp, /dev/null, /c/Users,
// bare /etc against the current drive) that the rest of the commands
// use. When a path behaves unexpectedly, namei shows exactly which
// component the translation produced and which one does not exist.
type nameiCommand struct{}

func (nameiCommand) Name() string    { return "namei" }
func (nameiCommand) Summary() string { return "follow a pathname until a terminal point is found" }

var nameiSpec = parser.Spec{
	{Short: 'l', Long: "long"},        // mode bits and owner
	{Short: 'm', Long: "modes"},       // mode bits
	{Short: 'o', Long: "owners"},      // owner
	{Short: 'x', Long: "mountpoints"}, // mark volume roots
}

func (nameiCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, nameiSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "namei: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: namei [-l] [-m] [-o] [-x] PATH...")
		return command.ExitUsage
	}

	long := res.Bool('l', "long")
	showModes := long || res.Bool('m', "modes")
	showOwner := long || res.Bool('o', "owners")
	markMounts := res.Bool('x', "mountpoints")

	exit := command.ExitSuccess
	for _, arg := range res.Positional {
		if !nameiWalk(ctx, arg, showModes, showOwner, markMounts) {
			exit = command.ExitFailure
		}
	}
	return exit
}

func nameiWalk(ctx *command.Context, arg string, showModes, showOwner, markMounts bool) bool {
	fmt.Fprintf(ctx.Stdout, "f: %s\n", arg)

	resolved, err := paths.Resolve(arg)
	if err != nil {
		fmt.Fprintf(ctx.Stdout, " ? %s - %s\n", arg, output.LinuxErrorText(err))
		return false
	}

	// The volume name is the Windows equivalent of "/" and is where the
	// translation is visible: "/c/Users/x" arrives here as "C:", "~" as
	// the drive holding the profile, "/etc" as the current drive.
	vol := filepath.VolumeName(resolved)
	root := vol + string(filepath.Separator)
	rest := strings.TrimPrefix(resolved, vol)
	rest = strings.Trim(rest, string(filepath.Separator))

	if !nameiPrint(ctx, root, root, showModes, showOwner, markMounts) {
		return false
	}

	current := root
	if rest == "" {
		return true
	}
	for _, comp := range strings.Split(rest, string(filepath.Separator)) {
		if comp == "" {
			continue
		}
		current = filepath.Join(current, comp)
		if !nameiPrint(ctx, current, comp, showModes, showOwner, markMounts) {
			// namei stops at the first component it cannot resolve,
			// since nothing below it can be examined.
			return false
		}
	}
	return true
}

// nameiPrint emits one component line and reports whether the walk can
// continue past it.
func nameiPrint(ctx *command.Context, path, label string, showModes, showOwner, markMounts bool) bool {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Fprintf(ctx.Stdout, " ? %s - %s\n", label, output.LinuxErrorText(err))
		return false
	}

	var fields []string
	if showModes {
		fields = append(fields, output.FormatMode(info))
	} else {
		fields = append(fields, string(nameiTypeChar(info, path, markMounts)))
	}
	if showOwner {
		// Windows security descriptors have no single "owner name" that
		// maps cleanly onto a Unix owner/group pair, so this follows the
		// same approximation ls -l uses across this project.
		fields = append(fields, output.CurrentUsername())
	}

	line := " " + strings.Join(fields, " ") + " " + label

	// A reparse point (symlink or directory junction) is the component
	// most likely to explain surprising behavior, so always show where
	// it points regardless of the verbosity flags.
	if info.Mode()&os.ModeSymlink != 0 {
		if target, err := os.Readlink(path); err == nil {
			line += " -> " + target
		}
	}
	fmt.Fprintln(ctx.Stdout, line)
	return true
}

func nameiTypeChar(info os.FileInfo, path string, markMounts bool) byte {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return 'l'
	case info.IsDir():
		// -x distinguishes a volume root, which is the closest Windows
		// equivalent of a mount point.
		if markMounts && filepath.Dir(path) == path {
			return 'D'
		}
		return 'd'
	default:
		return '-'
	}
}

func init() { command.Register(nameiCommand{}) }
