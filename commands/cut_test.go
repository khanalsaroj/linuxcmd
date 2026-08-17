package commands

import (
	"path/filepath"
	"testing"
)

func TestCutFieldsDefaultTabDelimiter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\tb\tc\n")

	code, out, errOut := run(t, "cut", "-f", "1,3", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\tc\n" {
		t.Errorf("cut -f 1,3 output = %q, want %q", out, "a\tc\n")
	}
}

func TestCutFieldsCustomDelimiter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a,b,c\n")

	code, out, errOut := run(t, "cut", "-d,", "-f2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "b\n" {
		t.Errorf("cut -d, -f2 output = %q, want %q", out, "b\n")
	}
}

func TestCutFieldsRange(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a,b,c,d\n")

	code, out, errOut := run(t, "cut", "-d,", "-f2-", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "b,c,d\n" {
		t.Errorf("cut -d, -f2- output = %q, want %q", out, "b,c,d\n")
	}
}

func TestCutCharacters(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abcdef\n")

	code, out, errOut := run(t, "cut", "-c1-3", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "abc\n" {
		t.Errorf("cut -c1-3 output = %q, want %q", out, "abc\n")
	}
}

func TestCutMissingSpecErrors(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\n")

	code, _, errOut := run(t, "cut", f)
	if code == 0 {
		t.Error("expected nonzero exit when neither -f nor -c is given")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
