package commands

import (
	"path/filepath"
	"testing"
)

func TestSedSubstituteGlobal(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "foo bar foo\n")

	code, out, errOut := run(t, "sed", "s/foo/bar/g", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "bar bar bar\n" {
		t.Errorf("sed output = %q, want %q", out, "bar bar bar\n")
	}
}

func TestSedSubstituteFirstOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "foo foo foo\n")

	code, out, errOut := run(t, "sed", "s/foo/bar/", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "bar foo foo\n" {
		t.Errorf("sed output = %q, want %q", out, "bar foo foo\n")
	}
}

func TestSedDelete(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "keep\nremove\nkeep2\n")

	code, out, errOut := run(t, "sed", "/remove/d", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "keep\nkeep2\n" {
		t.Errorf("sed output = %q, want %q", out, "keep\nkeep2\n")
	}
}

func TestSedBackreference(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "John Smith\n")

	code, out, errOut := run(t, "sed", `s/(\w+) (\w+)/\2 \1/`, f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "Smith John\n" {
		t.Errorf("sed output = %q, want %q", out, "Smith John\n")
	}
}

func TestSedInvalidScript(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x\n")

	code, _, errOut := run(t, "sed", "z/foo/bar/", f)
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
