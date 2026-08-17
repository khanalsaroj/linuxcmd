// Package output formats command results the way Linux tools do: terse
// "prog: verb 'target': reason" error lines and ls-style listings,
// translating Go/Windows error values into the message a Linux user
// would expect while still preserving Windows-specific detail when the
// translation doesn't apply.
package output

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// LinuxErrorText maps a Go error to the short reason string GNU
// coreutils would print (e.g. "No such file or directory"). If no
// well-known mapping applies, it falls back to the underlying error's own
// message so Windows-specific detail is never hidden.
func LinuxErrorText(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, fs.ErrNotExist), os.IsNotExist(err):
		return "No such file or directory"
	case errors.Is(err, fs.ErrPermission), os.IsPermission(err):
		return "Permission denied"
	case errors.Is(err, fs.ErrExist), os.IsExist(err):
		return "File exists"
	}

	msg := underlyingMessage(err)
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "directory not empty"):
		return "Directory not empty"
	case strings.Contains(lower, "not a directory"):
		return "Not a directory"
	case strings.Contains(lower, "is a directory"):
		return "Is a directory"
	case strings.Contains(lower, "access is denied"):
		return "Permission denied"
	case strings.Contains(lower, "cannot find the path"),
		strings.Contains(lower, "cannot find the file"),
		strings.Contains(lower, "system cannot find"):
		return "No such file or directory"
	case strings.Contains(lower, "file exists"), strings.Contains(lower, "already exists"):
		return "File exists"
	case strings.Contains(lower, "being used by another process"):
		return "Resource busy"
	}
	// No known mapping: surface the original Windows/Go message so
	// diagnostic detail isn't lost.
	return msg
}

// underlyingMessage extracts the innermost, most specific message from
// wrapped errors like *fs.PathError ("open x: access is denied") by
// preferring err.Err's message when present.
func underlyingMessage(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) && pathErr.Err != nil {
		return pathErr.Err.Error()
	}
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) && linkErr.Err != nil {
		return linkErr.Err.Error()
	}
	return err.Error()
}

// Errorf writes a Linux-style diagnostic line to w:
//
//	prog: verb 'target': reason
//
// e.g. "rm: cannot remove 'test.txt': No such file or directory".
func Errorf(w io.Writer, prog, verb, target string, err error) {
	fmt.Fprintf(w, "%s: %s '%s': %s\n", prog, verb, target, LinuxErrorText(err))
}

// SimpleErrorf writes a Linux-style diagnostic line without a verb
// clause: "prog: target: reason", e.g. "cat: missing.txt: No such file or
// directory". Used by tools (cat, grep) whose real-world messages skip
// the "cannot open '...'" wrapper that ls/cp/mv/rm use.
func SimpleErrorf(w io.Writer, prog, target string, err error) {
	fmt.Fprintf(w, "%s: %s: %s\n", prog, target, LinuxErrorText(err))
}
