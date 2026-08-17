package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStatRegularFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello")

	code, out, errOut := run(t, "stat", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "regular file") {
		t.Errorf("expected 'regular file' in output, got %q", out)
	}
	if !strings.Contains(out, "Size: 5") {
		t.Errorf("expected size 5 in output, got %q", out)
	}
}

func TestStatDirectory(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := run(t, "stat", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "directory") {
		t.Errorf("expected 'directory' in output, got %q", out)
	}
}

func TestStatMissingFile(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "stat", filepath.Join(dir, "nope.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}
