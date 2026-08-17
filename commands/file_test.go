package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFileDetectsText(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello world, this is plain text\n")

	code, out, errOut := run(t, "file", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "ASCII text") {
		t.Errorf("expected ASCII text detection, got %q", out)
	}
}

func TestFileDetectsPNG(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.png")
	mustWriteFile(t, f, "\x89PNG\r\n\x1a\nrest of file")

	code, out, errOut := run(t, "file", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "PNG image data") {
		t.Errorf("expected PNG detection, got %q", out)
	}
}

func TestFileDetectsEmpty(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.txt")
	mustWriteFile(t, f, "")

	code, out, errOut := run(t, "file", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "empty") {
		t.Errorf("expected empty detection, got %q", out)
	}
}

func TestFileDetectsDirectory(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := run(t, "file", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "directory") {
		t.Errorf("expected directory detection, got %q", out)
	}
}
