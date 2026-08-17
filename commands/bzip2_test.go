package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBzip2Decompress(t *testing.T) {
	bzip2Exe, err := exec.LookPath("bzip2")
	if err != nil {
		t.Skip("no real bzip2 binary available to create a test fixture")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "data.txt")
	mustWriteFile(t, src, "hello bzip2 world\n")
	if err := exec.Command(bzip2Exe, "-k", src).Run(); err != nil {
		t.Skipf("bzip2 fixture creation failed: %v", err)
	}

	if err := os.Remove(src); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "bzip2", "-d", src+".bz2")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello bzip2 world\n" {
		t.Errorf("decompressed content = %q, want %q", got, "hello bzip2 world\n")
	}
}

func TestBzip2CompressionUnsupported(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "bzip2", f)
	if code == 0 {
		t.Error("expected nonzero exit since compression is unsupported")
	}
	if errOut == "" {
		t.Error("expected an explanatory error message")
	}
}

func TestXzReportsUnsupported(t *testing.T) {
	code, _, errOut := run(t, "xz", "-d", "archive.xz")
	if code == 0 {
		t.Error("expected nonzero exit since xz is unsupported")
	}
	if errOut == "" {
		t.Error("expected an explanatory error message")
	}
}
