package commands

import (
	"bytes"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

func TestHeadDefaultTenLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	var lines []string
	for i := 1; i <= 20; i++ {
		lines = append(lines, "line"+strconv.Itoa(i))
	}
	mustWriteFile(t, f, strings.Join(lines, "\n")+"\n")

	code, out, errOut := run(t, "head", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(got) != 10 || got[0] != "line1" || got[9] != "line10" {
		t.Errorf("head output = %q", out)
	}
}

func TestHeadNLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\nd\n")

	code, out, errOut := run(t, "head", "-n", "2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\nb\n" {
		t.Errorf("head -n 2 output = %q, want %q", out, "a\nb\n")
	}
}

func TestHeadBytes(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abcdef")

	code, out, errOut := run(t, "head", "-c", "3", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "abc" {
		t.Errorf("head -c 3 output = %q, want %q", out, "abc")
	}
}

func TestHeadFewerLinesThanRequested(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\n")

	code, out, errOut := run(t, "head", "-n", "10", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\nb\n" {
		t.Errorf("head output = %q, want %q", out, "a\nb\n")
	}
}

func TestHeadMultipleFilesShowsHeaders(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "AAA\n")
	mustWriteFile(t, b, "BBB\n")

	code, out, errOut := run(t, "head", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "==> " + a + " <==\nAAA\n\n==> " + b + " <==\nBBB\n"
	if out != want {
		t.Errorf("head output = %q, want %q", out, want)
	}
}

func TestHeadSingleFileNoHeader(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "AAA\n")

	code, out, errOut := run(t, "head", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "==>") {
		t.Errorf("expected no header for a single file, got %q", out)
	}
}

func TestHeadQuietSuppressesHeaders(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "AAA\n")
	mustWriteFile(t, b, "BBB\n")

	code, out, errOut := run(t, "head", "-q", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "==>") {
		t.Errorf("expected -q to suppress headers, got %q", out)
	}
}

func TestHeadStdin(t *testing.T) {
	cmd, ok := command.Lookup("head")
	if !ok {
		t.Fatal("head is not registered")
	}
	var outBuf, errBuf bytes.Buffer
	ctx := &command.Context{
		CommandName: "head",
		Args:        []string{"-n", "2"},
		Stdin:       strings.NewReader("x\ny\nz\n"),
		Stdout:      &outBuf,
		Stderr:      &errBuf,
	}
	if code := cmd.Run(ctx); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errBuf.String())
	}
	if outBuf.String() != "x\ny\n" {
		t.Errorf("head stdin output = %q, want %q", outBuf.String(), "x\ny\n")
	}
}

func TestHeadMissingFile(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "head", filepath.Join(dir, "nope.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestHeadDirectory(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "head", dir)
	if code == 0 {
		t.Error("expected nonzero exit for a directory")
	}
	if !strings.Contains(errOut, "Is a directory") {
		t.Errorf("expected 'Is a directory' error, got %q", errOut)
	}
}

func TestHeadInvalidLineCount(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\n")

	code, _, errOut := run(t, "head", "-n", "notanumber", f)
	if code == 0 {
		t.Error("expected nonzero exit for invalid -n value")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
