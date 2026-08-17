package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTeeWritesStdoutAndFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")

	code, out, errOut := runWithStdin(t, "tee", "hello\n", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello\n" {
		t.Errorf("tee stdout = %q, want %q", out, "hello\n")
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello\n" {
		t.Errorf("tee file content = %q, want %q", got, "hello\n")
	}
}

func TestTeeAppend(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")
	mustWriteFile(t, f, "first\n")

	code, _, errOut := runWithStdin(t, "tee", "second\n", "-a", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first\nsecond\n" {
		t.Errorf("tee -a file content = %q, want %q", got, "first\nsecond\n")
	}
}

func TestTeeTruncatesByDefault(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")
	mustWriteFile(t, f, "old content that is longer\n")

	code, _, errOut := runWithStdin(t, "tee", "new\n", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new\n" {
		t.Errorf("tee file content = %q, want %q", got, "new\n")
	}
}
