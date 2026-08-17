package commands

import (
	"path/filepath"
	"testing"
)

func TestSortAlphabetic(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "banana\napple\ncherry\n")

	code, out, errOut := run(t, "sort", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "apple\nbanana\ncherry\n"
	if out != want {
		t.Errorf("sort output = %q, want %q", out, want)
	}
}

func TestSortNumeric(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "10\n2\n1\n")

	code, out, errOut := run(t, "sort", "-n", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "1\n2\n10\n"
	if out != want {
		t.Errorf("sort -n output = %q, want %q", out, want)
	}
}

func TestSortReverse(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nc\nb\n")

	code, out, errOut := run(t, "sort", "-r", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "c\nb\na\n"
	if out != want {
		t.Errorf("sort -r output = %q, want %q", out, want)
	}
}

func TestSortUnique(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "b\na\nb\na\n")

	code, out, errOut := run(t, "sort", "-u", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a\nb\n"
	if out != want {
		t.Errorf("sort -u output = %q, want %q", out, want)
	}
}

func TestSortByKey(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "b 2\na 1\nc 3\n")

	code, out, errOut := run(t, "sort", "-k", "2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a 1\nb 2\nc 3\n"
	if out != want {
		t.Errorf("sort -k 2 output = %q, want %q", out, want)
	}
}
