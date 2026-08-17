package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDdCopiesFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.bin")
	dst := filepath.Join(dir, "out.bin")
	mustWriteFile(t, src, strings.Repeat("x", 2048))

	code, _, errOut := run(t, "dd", "if="+src, "of="+dst, "bs=512")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(errOut, "records in") {
		t.Errorf("expected a records summary, got %q", errOut)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2048 {
		t.Errorf("copied %d bytes, want 2048", len(got))
	}
}

func TestDdCount(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.bin")
	dst := filepath.Join(dir, "out.bin")
	mustWriteFile(t, src, strings.Repeat("y", 4096))

	code, _, errOut := run(t, "dd", "if="+src, "of="+dst, "bs=512", "count=2")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1024 {
		t.Errorf("copied %d bytes, want 1024", len(got))
	}
}

func TestDdInvalidOperand(t *testing.T) {
	code, _, errOut := run(t, "dd", "bogus")
	if code == 0 {
		t.Error("expected nonzero exit for an unrecognized operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
