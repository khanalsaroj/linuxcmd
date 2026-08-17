package commands

import (
	"path/filepath"
	"testing"
)

func TestMoreStreamsWhenNotInteractive(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "line1\nline2\n")

	code, out, errOut := run(t, "more", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "line1\nline2\n" {
		t.Errorf("more output = %q, want %q", out, "line1\nline2\n")
	}
}

func TestLessStreamsWhenNotInteractive(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello\n")

	code, out, errOut := run(t, "less", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello\n" {
		t.Errorf("less output = %q, want %q", out, "hello\n")
	}
}
