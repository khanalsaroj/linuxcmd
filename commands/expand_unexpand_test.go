package commands

import (
	"path/filepath"
	"testing"
)

func TestExpandConvertsTabsToSpaces(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\tb\n")

	code, out, errOut := run(t, "expand", "-t", "4", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a   b\n"
	if out != want {
		t.Errorf("expand output = %q, want %q", out, want)
	}
}

func TestUnexpandConvertsLeadingSpacesToTabs(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "    a\n")

	code, out, errOut := run(t, "unexpand", "-t", "4", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "\ta\n"
	if out != want {
		t.Errorf("unexpand output = %q, want %q", out, want)
	}
}

func TestUnexpandOnlyLeadingByDefault(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "    a    b\n")

	code, out, errOut := run(t, "unexpand", "-t", "4", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "\ta    b\n"
	if out != want {
		t.Errorf("unexpand output = %q, want %q", out, want)
	}
}
