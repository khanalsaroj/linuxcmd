package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWCDefaultCounts(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "one two\nthree\n")

	code, out, errOut := run(t, "wc", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	fields := strings.Fields(out)
	if len(fields) != 4 || fields[0] != "2" || fields[1] != "3" || fields[2] != "14" || fields[3] != f {
		t.Errorf("wc output = %q", out)
	}
}

func TestWCLinesOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\n")

	code, out, errOut := run(t, "wc", "-l", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	fields := strings.Fields(out)
	if len(fields) != 2 || fields[0] != "3" {
		t.Errorf("wc -l output = %q", out)
	}
}

func TestWCBytes(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abcde")

	code, out, errOut := run(t, "wc", "-c", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	fields := strings.Fields(out)
	if len(fields) != 2 || fields[0] != "5" {
		t.Errorf("wc -c output = %q", out)
	}
}

func TestWCMultipleFilesTotal(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "a\nb\n")
	mustWriteFile(t, b, "c\n")

	code, out, errOut := run(t, "wc", "-l", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 output lines (a, b, total), got %q", out)
	}
	if !strings.Contains(lines[2], "total") {
		t.Errorf("expected a total line, got %q", lines[2])
	}
}

func TestWCMissingFile(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "wc", filepath.Join(dir, "nope.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}
