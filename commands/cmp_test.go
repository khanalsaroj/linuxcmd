package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCmpIdenticalFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "same content")
	mustWriteFile(t, b, "same content")

	code, out, errOut := run(t, "cmp", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "" {
		t.Errorf("expected no output for identical files, got %q", out)
	}
}

func TestCmpDifferingFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "hello")
	mustWriteFile(t, b, "hallo")

	code, out, _ := run(t, "cmp", a, b)
	if code == 0 {
		t.Error("expected nonzero exit for differing files")
	}
	if !strings.Contains(out, "byte 2") {
		t.Errorf("expected first difference at byte 2, got %q", out)
	}
}

func TestCmpDifferingLength(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "short")
	mustWriteFile(t, b, "shortlonger")

	code, out, _ := run(t, "cmp", a, b)
	if code == 0 {
		t.Error("expected nonzero exit for differing lengths")
	}
	if !strings.Contains(out, "EOF") {
		t.Errorf("expected EOF message, got %q", out)
	}
}
