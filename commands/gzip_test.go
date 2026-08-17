package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGzipCompressAndDecompress(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello gzip world")

	code, _, errOut := run(t, "gzip", f)
	if code != 0 {
		t.Fatalf("gzip exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("expected original file to be removed after gzip")
	}
	gz := f + ".gz"
	if _, err := os.Stat(gz); err != nil {
		t.Fatalf("expected %s to exist: %v", gz, err)
	}

	code, _, errOut = run(t, "gzip", "-d", gz)
	if code != 0 {
		t.Fatalf("gzip -d exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello gzip world" {
		t.Errorf("decompressed content = %q, want %q", got, "hello gzip world")
	}
}

func TestGzipKeepFlag(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "keep me")

	code, _, errOut := run(t, "gzip", "-k", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); err != nil {
		t.Error("expected original file to remain with -k")
	}
}
