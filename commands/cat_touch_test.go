package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCATSingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "line1\nline2\n")

	code, out, errOut := run(t, "cat", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "line1\nline2\n" {
		t.Errorf("cat output = %q, want %q", out, "line1\nline2\n")
	}
}

func TestCATConcatenatesMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "AAA")
	mustWriteFile(t, b, "BBB")

	code, out, errOut := run(t, "cat", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "AAABBB" {
		t.Errorf("cat output = %q, want %q", out, "AAABBB")
	}
}

func TestCATMissingFile(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "cat", filepath.Join(dir, "nope.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestCATDirectory(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "cat", dir)
	if code == 0 {
		t.Error("expected nonzero exit for a directory")
	}
	if !strings.Contains(errOut, "Is a directory") {
		t.Errorf("expected 'Is a directory' error, got %q", errOut)
	}
}

func TestCATNumberedLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\n")

	code, out, errOut := run(t, "cat", "-n", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "1\ta") || !strings.Contains(out, "2\tb") {
		t.Errorf("expected numbered lines, got %q", out)
	}
}

func TestTouchCreatesFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "new.txt")

	code, _, errOut := run(t, "touch", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); err != nil {
		t.Errorf("expected touch to create %q", f)
	}
}

func TestTouchUpdatesExistingTimestamp(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "existing.txt")
	mustWriteFile(t, f, "x")
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(f, old, old); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "touch", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(info.ModTime()) > time.Minute {
		t.Errorf("expected touch to refresh mtime, still shows %v", info.ModTime())
	}
}

func TestTouchNoCreateFlag(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "absent.txt")

	code, _, errOut := run(t, "touch", "-c", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("expected touch -c to not create a missing file")
	}
}

func TestTouchMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "touch")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
