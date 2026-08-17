package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDuSummarize(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), strings.Repeat("x", 1000))
	mustWriteFile(t, filepath.Join(dir, "sub", "b.txt"), strings.Repeat("y", 1000))

	code, out, errOut := run(t, "du", "-s", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected a single summarized line, got %q", out)
	}
	if !strings.Contains(lines[0], dir) {
		t.Errorf("expected output to reference %q, got %q", dir, lines[0])
	}
}

func TestDuSingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	mustWriteFile(t, f, "hello")

	code, out, errOut := run(t, "du", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, f) {
		t.Errorf("expected output to reference %q, got %q", f, out)
	}
}

func TestDuHumanReadable(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), strings.Repeat("x", 2000))

	code, out, errOut := run(t, "du", "-sh", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty output")
	}
}
