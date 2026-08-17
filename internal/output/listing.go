package output

import (
	"fmt"
	"io/fs"
	"os/user"
	"strings"
	"sync"
	"time"
)

// LongEntry holds the fields needed to render one "ls -l" line.
type LongEntry struct {
	Info       fs.FileInfo
	LinkTarget string // set if the entry is a symlink; "" otherwise
}

var (
	currentUserOnce sync.Once
	currentUserName = "user"
)

// CurrentUsername returns the invoking user's name with any DOMAIN\ or
// COMPUTERNAME\ prefix stripped, matching the bare username Linux tools
// display. Used by both "ls -l"'s owner column and "whoami".
func CurrentUsername() string {
	currentUserOnce.Do(func() {
		if u, err := user.Current(); err == nil && u.Username != "" {
			name := u.Username
			// Strip a DOMAIN\ or COMPUTER\ prefix Windows adds, to keep
			// the column narrow like a Unix username would be.
			if idx := strings.LastIndexByte(name, '\\'); idx != -1 {
				name = name[idx+1:]
			}
			currentUserName = name
		}
	})
	return currentUserName
}

// FormatMode renders a Linux-style 10-character mode string (e.g.
// "drwxr-xr-x") from Windows file attributes. Windows has no POSIX
// permission bits, so this is a best-effort approximation documented as a
// limitation in the README: directories are always traversable, files are
// writable unless the read-only attribute is set, and there is no
// concept of per-class (owner/group/other) permissions, so all three
// triplets are identical.
func FormatMode(info fs.FileInfo) string {
	var b strings.Builder
	mode := info.Mode()

	switch {
	case mode&fs.ModeSymlink != 0:
		b.WriteByte('l')
	case info.IsDir():
		b.WriteByte('d')
	default:
		b.WriteByte('-')
	}

	readOnly := mode&0200 == 0 // Go maps the Windows read-only attribute onto the owner-write bit
	var triplet string
	switch {
	case info.IsDir():
		triplet = "rwx" // directories are always traversable
	case readOnly:
		triplet = "r--"
	default:
		triplet = "rw-"
	}
	b.WriteString(triplet)
	b.WriteString(triplet)
	b.WriteString(triplet)
	return b.String()
}

// FormatLongLine renders one entry the way "ls -l" would, e.g.:
//
//	drwxr-xr-x  1 user  group      4096 Aug 17 12:30 Documents
func FormatLongLine(e LongEntry) string {
	name := e.Info.Name()
	if e.LinkTarget != "" {
		name = name + " -> " + e.LinkTarget
	}
	modTime := e.Info.ModTime()
	when := formatTime(modTime)
	owner := CurrentUsername()
	return fmt.Sprintf("%s %3d %-8s %-8s %10d %s %s",
		FormatMode(e.Info), 1, owner, owner, e.Info.Size(), when, name)
}

// formatTime mimics GNU ls: "Mon _2 15:04" for files modified within the
// last ~6 months, "Mon _2  2006" (year instead of time) for older files.
func formatTime(t time.Time) string {
	if time.Since(t) > 182*24*time.Hour {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}
