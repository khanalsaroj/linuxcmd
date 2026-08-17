// Package fsutil provides filesystem operations shared by multiple
// commands (cp, mv, rm) that need more than a single os.* call: recursive
// copy/remove, and move that falls back to copy+delete across drives
// since os.Rename cannot cross volumes on Windows.
package fsutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyFile copies a single file's contents and mode. It does not follow
// symlinks specially; Go's os.Open already dereferences them.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// CopyRecursive copies src to dst. If src is a directory, dst is created
// and the tree is copied into it recursively.
func CopyRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return CopyFile(src, dst)
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := CopyRecursive(s, d); err != nil {
			return err
		}
	}
	return nil
}

// Move renames src to dst, falling back to copy-then-delete when the
// paths span different volumes (os.Rename fails with a cross-device
// error in that case, unlike Linux where it can fall back to the same
// behavior automatically for some filesystems).
func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if !isCrossDevice(err) {
		return err
	}
	if err := CopyRecursive(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func isCrossDevice(err error) bool {
	// syscall.ERROR_NOT_SAME_DEVICE (Windows) surfaces through
	// LinkError/PathError with this text; matching on it avoids an extra
	// build-tag-gated import of the exact syscall type.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not the same") || strings.Contains(msg, "different disk") || strings.Contains(msg, "not on the same")
}

// RemoveRecursive removes path, whether file or directory tree.
func RemoveRecursive(path string) error {
	return os.RemoveAll(path)
}
