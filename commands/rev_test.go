package commands

import (
	"path/filepath"
	"testing"
)

func TestRevReversesEachLine(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abc\nxyz\n")

	code, out, errOut := run(t, "rev", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "cba\nzyx\n"
	if out != want {
		t.Errorf("rev output = %q, want %q", out, want)
	}
}
