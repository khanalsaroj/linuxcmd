package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShredOverwritesFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "secret.txt")
	mustWriteFile(t, f, "sensitive content")

	code, _, errOut := run(t, "shred", "-n", "1", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "sensitive content" {
		t.Error("expected file content to be overwritten")
	}
	if _, err := os.Stat(f); err != nil {
		t.Error("expected file to still exist without -u")
	}
}

func TestShredRemovesWithU(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "secret.txt")
	mustWriteFile(t, f, "sensitive content")

	code, _, errOut := run(t, "shred", "-u", "-n", "1", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("expected file to be removed with -u")
	}
}

func TestShredMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "shred")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
