package commands

import (
	"strings"
	"testing"
)

func TestPWD(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := runIn(t, dir, "pwd")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.TrimSpace(out)
	if !strings.EqualFold(got, dir) {
		t.Errorf("pwd output %q, want %q (case-insensitive)", got, dir)
	}
}

func TestCDResolvesHome(t *testing.T) {
	code, out, errOut := run(t, "cd", "~")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected cd ~ to print a resolved path")
	}
}

func TestCDMissingDirectory(t *testing.T) {
	code, _, errOut := run(t, "cd", "this-directory-does-not-exist-xyz")
	if code == 0 {
		t.Error("expected nonzero exit for missing directory")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestCDTargetIsAFile(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\f.txt`, "x")

	code, _, errOut := run(t, "cd", dir+`\f.txt`)
	if code == 0 {
		t.Error("expected nonzero exit when target is a file")
	}
	if !strings.Contains(errOut, "Not a directory") {
		t.Errorf("expected 'Not a directory' error, got %q", errOut)
	}
}

func TestCDValidDirectory(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := run(t, "cd", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected resolved path in output")
	}
}
