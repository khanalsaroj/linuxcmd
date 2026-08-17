// Package paths normalizes Linux-style path syntax into real Windows
// paths. It never assumes "/" is the Windows filesystem root: Windows has
// no single root shared by all drives, so a bare Linux-style absolute path
// is resolved relative to the current drive's root, and well-known Linux
// paths (~, /tmp, /dev/null) are mapped to their closest Windows
// equivalent. Everything else passes through with slashes converted to
// the native separator.
package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Normalize converts a user-supplied path (which may use Linux syntax:
// forward slashes, ~, /tmp, /dev/null, /c/... drive shorthand) into a
// native Windows path. It does not make the result absolute or check
// existence; call Resolve for that.
func Normalize(p string) string {
	if p == "" {
		return p
	}

	// Expand a leading ~ to the user's home directory. "~" alone, "~/rest"
	// and "~\rest" are all recognized; "~user" (another user's home) is
	// not supported since Windows has no equivalent concept exposed here.
	if p == "~" {
		return home()
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		return filepath.Join(home(), p[2:])
	}

	// Well-known Linux paths with a specific Windows equivalent.
	switch {
	case p == "/dev/null":
		return os.DevNull
	case p == "/tmp" || strings.HasPrefix(p, "/tmp/"):
		return filepath.Join(os.TempDir(), strings.TrimPrefix(p, "/tmp"))
	}

	// Windows-style absolute path (C:\..., C:/..., or UNC \\server\share)
	// or a relative path: just normalize slash direction.
	if isWindowsAbs(p) || strings.HasPrefix(p, `\\`) || !strings.HasPrefix(p, "/") {
		return filepath.FromSlash(p)
	}

	// From here p starts with a single "/". Recognize the MSYS/Git-Bash
	// drive-letter convention "/c/Users/..." -> "C:\Users\...".
	rest := p[1:]
	if len(rest) >= 1 {
		letter := rest[0]
		isDriveLetter := (letter >= 'a' && letter <= 'z') || (letter >= 'A' && letter <= 'Z')
		if isDriveLetter && (len(rest) == 1 || rest[1] == '/') {
			driveRest := rest[1:]
			return filepath.FromSlash(strings.ToUpper(string(letter)) + ":" + driveRest)
		}
	}

	// A bare Linux-style absolute path with no Windows equivalent:
	// resolve it against the root of the current working directory's
	// drive, e.g. "/etc" from C:\Users\foo becomes "C:\etc". This is the
	// closest meaningful representation Windows offers, since there is no
	// single filesystem root shared across drives.
	vol := filepath.VolumeName(cwd())
	if vol == "" {
		vol = "C:"
	}
	return filepath.FromSlash(vol + p)
}

// Resolve normalizes p and makes it absolute against the current working
// directory, then cleans it (collapsing "." and "..").
func Resolve(p string) (string, error) {
	n := Normalize(p)
	if !filepath.IsAbs(n) {
		abs, err := filepath.Abs(n)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return filepath.Clean(n), nil
}

func isWindowsAbs(p string) bool {
	if len(p) >= 3 {
		c := p[0]
		if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) && p[1] == ':' && (p[2] == '/' || p[2] == '\\') {
			return true
		}
	}
	return false
}

func home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return `C:\`
	}
	return h
}

func cwd() string {
	d, err := os.Getwd()
	if err != nil {
		return `C:\`
	}
	return d
}

// Display converts a native Windows path to forward-slash form for
// Linux-style display purposes (used by commands whose output looks more
// natural with "/" separators, e.g. echoing back a path the user typed
// with forward slashes). It does not change drive letters or roots.
func Display(p string) string {
	return filepath.ToSlash(p)
}

// Home returns the user's home directory, used for "cd" with no
// arguments and for "~" expansion.
func Home() string {
	return home()
}
