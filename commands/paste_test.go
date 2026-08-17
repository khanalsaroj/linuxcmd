package commands

import (
	"path/filepath"
	"testing"
)

func TestPasteMergesLines(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "names.txt")
	b := filepath.Join(dir, "ids.txt")
	mustWriteFile(t, a, "alice\nbob\n")
	mustWriteFile(t, b, "1\n2\n")

	code, out, errOut := run(t, "paste", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "alice\t1\nbob\t2\n"
	if out != want {
		t.Errorf("paste output = %q, want %q", out, want)
	}
}

func TestPasteCustomDelimiter(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "x\n")
	mustWriteFile(t, b, "y\n")

	code, out, errOut := run(t, "paste", "-d,", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "x,y\n" {
		t.Errorf("paste -d, output = %q, want %q", out, "x,y\n")
	}
}
