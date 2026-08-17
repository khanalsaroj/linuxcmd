package commands

import (
	"path/filepath"
	"testing"
)

func TestUniqCollapsesAdjacentDuplicates(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\na\nb\nb\nb\nc\n")

	code, out, errOut := run(t, "uniq", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a\nb\nc\n"
	if out != want {
		t.Errorf("uniq output = %q, want %q", out, want)
	}
}

func TestUniqCount(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\na\nb\n")

	code, out, errOut := run(t, "uniq", "-c", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "      2 a\n      1 b\n"
	if out != want {
		t.Errorf("uniq -c output = %q, want %q", out, want)
	}
}

func TestUniqNonAdjacentDuplicatesKept(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\na\n")

	code, out, errOut := run(t, "uniq", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a\nb\na\n"
	if out != want {
		t.Errorf("uniq output = %q, want %q", out, want)
	}
}

func TestUniqRepeatedOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\na\nb\nc\nc\n")

	code, out, errOut := run(t, "uniq", "-d", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a\nc\n"
	if out != want {
		t.Errorf("uniq -d output = %q, want %q", out, want)
	}
}
