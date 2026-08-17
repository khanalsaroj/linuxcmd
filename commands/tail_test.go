package commands

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTailDefaultTenLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	var b strings.Builder
	for i := 1; i <= 20; i++ {
		b.WriteString("line")
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('\n')
	}
	mustWriteFile(t, f, b.String())

	code, out, errOut := run(t, "tail", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(got) != 10 || got[0] != "line11" || got[9] != "line20" {
		t.Errorf("tail output = %q", out)
	}
}

func TestTailNLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\nd\n")

	code, out, errOut := run(t, "tail", "-n", "2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "c\nd\n" {
		t.Errorf("tail -n 2 output = %q, want %q", out, "c\nd\n")
	}
}

func TestTailBytes(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abcdef")

	code, out, errOut := run(t, "tail", "-c", "3", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "def" {
		t.Errorf("tail -c 3 output = %q, want %q", out, "def")
	}
}

func TestTailFewerLinesThanRequested(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\n")

	code, out, errOut := run(t, "tail", "-n", "10", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\nb\n" {
		t.Errorf("tail output = %q, want %q", out, "a\nb\n")
	}
}

func TestTailNoTrailingNewlinePreserved(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc")

	code, out, errOut := run(t, "tail", "-n", "2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "b\nc" {
		t.Errorf("tail output = %q, want %q", out, "b\nc")
	}
}

func TestTailMultipleFilesShowsHeaders(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "AAA\n")
	mustWriteFile(t, b, "BBB\n")

	code, out, errOut := run(t, "tail", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "==> " + a + " <==\nAAA\n\n==> " + b + " <==\nBBB\n"
	if out != want {
		t.Errorf("tail output = %q, want %q", out, want)
	}
}

func TestTailFollowRejectsMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "AAA\n")
	mustWriteFile(t, b, "BBB\n")

	code, _, errOut := run(t, "tail", "-f", a, b)
	if code == 0 {
		t.Error("expected nonzero exit for -f with multiple files")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestTailMissingFile(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "tail", filepath.Join(dir, "nope.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestLastLinesHelper(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"a\nb\nc\nd\n", 2, "c\nd\n"},
		{"a\nb\nc\nd\n", 10, "a\nb\nc\nd\n"},
		{"a\nb\nc", 2, "b\nc"},
		{"", 5, ""},
	}
	for _, c := range cases {
		got := string(lastLines([]byte(c.in), c.n))
		if got != c.want {
			t.Errorf("lastLines(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}
