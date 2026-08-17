package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChmodOctalReadOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "chmod", "444", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0200 != 0 {
		t.Error("expected file to be read-only after chmod 444")
	}
}

func TestChmodMinusW(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "chmod", "-w", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0200 != 0 {
		t.Error("expected file to be read-only after chmod -w")
	}

	// Restore write permission so t.TempDir() cleanup can remove it.
	code, _, errOut = run(t, "chmod", "+w", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}

func TestChmodInvalidMode(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "chmod", "bogus", f)
	if code == 0 {
		t.Error("expected nonzero exit for an invalid mode")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
