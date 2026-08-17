package commands

import (
	"path/filepath"
	"testing"
)

func TestAwkPrintField(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "alice 30\nbob 25\n")

	code, out, errOut := run(t, "awk", "{print $1}", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "alice\nbob\n" {
		t.Errorf("awk output = %q, want %q", out, "alice\nbob\n")
	}
}

func TestAwkPrintMultipleFields(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a b c\n")

	code, out, errOut := run(t, "awk", "{print $2, $1}", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "b a\n" {
		t.Errorf("awk output = %q, want %q", out, "b a\n")
	}
}

func TestAwkPattern(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "keep me\nskip this\nkeep too\n")

	code, out, errOut := run(t, "awk", "/keep/", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "keep me\nkeep too\n" {
		t.Errorf("awk output = %q, want %q", out, "keep me\nkeep too\n")
	}
}

func TestAwkCustomFieldSeparator(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a,b,c\n")

	code, out, errOut := run(t, "awk", "-F", ",", "{print $2}", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "b\n" {
		t.Errorf("awk output = %q, want %q", out, "b\n")
	}
}

func TestAwkNF(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a b c d\n")

	code, out, errOut := run(t, "awk", "{print $NF}", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "d\n" {
		t.Errorf("awk output = %q, want %q", out, "d\n")
	}
}

func TestAwkUnsupportedAction(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x\n")

	code, _, errOut := run(t, "awk", "{x=1}", f)
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported action")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
