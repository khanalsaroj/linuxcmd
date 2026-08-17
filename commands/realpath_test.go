package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRealpathCleansRelativePath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sub", "..", "f.txt")
	mustWriteFile(t, filepath.Join(dir, "f.txt"), "x")

	code, out, errOut := run(t, "realpath", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := filepath.Join(dir, "f.txt")
	if strings.TrimSpace(out) != want {
		t.Errorf("realpath output = %q, want %q", strings.TrimSpace(out), want)
	}
}

func TestRealpathMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "realpath")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
