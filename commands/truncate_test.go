package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTruncateShrinks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello world")

	code, _, errOut := run(t, "truncate", "-s", "5", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("truncated content = %q, want %q", got, "hello")
	}
}

func TestTruncateCreatesMissingFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "new.txt")

	code, _, errOut := run(t, "truncate", "-s", "0", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); err != nil {
		t.Errorf("expected truncate to create %q: %v", f, err)
	}
}
