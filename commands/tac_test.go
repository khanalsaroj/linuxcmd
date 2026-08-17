package commands

import (
	"path/filepath"
	"testing"
)

func TestTacReversesLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\n")

	code, out, errOut := run(t, "tac", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "c\nb\na\n"
	if out != want {
		t.Errorf("tac output = %q, want %q", out, want)
	}
}
