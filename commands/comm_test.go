package commands

import (
	"path/filepath"
	"testing"
)

func TestCommThreeColumns(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "apple\nbanana\ncherry\n")
	mustWriteFile(t, b, "banana\ncherry\ndate\n")

	code, out, errOut := run(t, "comm", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "apple\n\t\tbanana\n\t\tcherry\n\tdate\n"
	if out != want {
		t.Errorf("comm output = %q, want %q", out, want)
	}
}

func TestCommSuppressColumns(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "apple\nbanana\n")
	mustWriteFile(t, b, "banana\ndate\n")

	code, out, errOut := run(t, "comm", "-1", "-2", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "banana\n"
	if out != want {
		t.Errorf("comm -1 -2 output = %q, want %q", out, want)
	}
}
