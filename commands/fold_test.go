package commands

import (
	"path/filepath"
	"testing"
)

func TestFoldWrapsLongLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abcdefghij\n")

	code, out, errOut := run(t, "fold", "-w", "4", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "abcd\nefgh\nij\n"
	if out != want {
		t.Errorf("fold output = %q, want %q", out, want)
	}
}

func TestFoldShortLineUnchanged(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abc\n")

	code, out, errOut := run(t, "fold", "-w", "80", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "abc\n" {
		t.Errorf("fold output = %q, want %q", out, "abc\n")
	}
}
