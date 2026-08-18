package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDos2unixConvertsInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crlf.txt")
	mustWriteFile(t, path, "one\r\ntwo\r\n")

	code, _, errOut := run(t, "dos2unix", "-q", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := mustReadFile(t, path)
	if got != "one\ntwo\n" {
		t.Errorf("converted content = %q, want %q", got, "one\ntwo\n")
	}
}

func TestUnix2dosConvertsInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lf.txt")
	mustWriteFile(t, path, "one\ntwo\n")

	code, _, errOut := run(t, "unix2dos", "-q", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := mustReadFile(t, path)
	if got != "one\r\ntwo\r\n" {
		t.Errorf("converted content = %q, want %q", got, "one\r\ntwo\r\n")
	}
}

// unix2dos must not turn an already-converted file into CR CR LF, which
// is why the implementation normalizes to LF before expanding.
func TestUnix2dosIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "already.txt")
	mustWriteFile(t, path, "a\r\nb\r\n")

	for i := 0; i < 3; i++ {
		if code, _, errOut := run(t, "unix2dos", "-q", path); code != 0 {
			t.Fatalf("run %d: exit code = %d, stderr = %q", i, code, errOut)
		}
	}
	if got := mustReadFile(t, path); got != "a\r\nb\r\n" {
		t.Errorf("content after repeated runs = %q, want %q", got, "a\r\nb\r\n")
	}
}

func TestDos2unixRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "round.txt")
	original := "x\r\ny\r\nz\r\n"
	mustWriteFile(t, path, original)

	if code, _, errOut := run(t, "dos2unix", "-q", path); code != 0 {
		t.Fatalf("dos2unix: exit code = %d, stderr = %q", code, errOut)
	}
	if code, _, errOut := run(t, "unix2dos", "-q", path); code != 0 {
		t.Fatalf("unix2dos: exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, path); got != original {
		t.Errorf("round trip = %q, want %q", got, original)
	}
}

// A file containing a NUL is binary; rewriting its CR bytes would
// corrupt it, so it must be skipped without being reported as an error.
func TestDos2unixSkipsBinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.dat")
	original := "a\r\n\x00b\r\n"
	mustWriteFile(t, path, original)

	code, _, errOut := run(t, "dos2unix", path)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 for a skipped binary file", code)
	}
	if !strings.Contains(errOut, "Skipping binary file") {
		t.Errorf("stderr = %q, want a skip notice", errOut)
	}
	if got := mustReadFile(t, path); got != original {
		t.Errorf("binary file was modified: %q", got)
	}
}

func TestDos2unixForcesBinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.dat")
	mustWriteFile(t, path, "a\r\n\x00b\r\n")

	if code, _, errOut := run(t, "dos2unix", "-q", "-f", path); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, path); got != "a\n\x00b\n" {
		t.Errorf("forced conversion = %q", got)
	}
}

func TestDos2unixFiltersStdin(t *testing.T) {
	code, out, errOut := runWithStdin(t, "dos2unix", "p\r\nq\r\n")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "p\nq\n" {
		t.Errorf("stdin filter = %q, want %q", out, "p\nq\n")
	}
}

func TestDos2unixNewFileLeavesSourceAlone(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.txt")
	dst := filepath.Join(dir, "out.txt")
	mustWriteFile(t, src, "k\r\n")

	if code, _, errOut := run(t, "dos2unix", "-q", "-n", src, dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, src); got != "k\r\n" {
		t.Errorf("source was modified: %q", got)
	}
	if got := mustReadFile(t, dst); got != "k\n" {
		t.Errorf("output file = %q, want %q", got, "k\n")
	}
}

func TestDos2unixReportsMissingFile(t *testing.T) {
	code, _, errOut := run(t, "dos2unix", filepath.Join(t.TempDir(), "absent.txt"))
	if code == 0 {
		t.Error("expected a nonzero exit for a missing file")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("stderr = %q, want a Linux-style not-found message", errOut)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
