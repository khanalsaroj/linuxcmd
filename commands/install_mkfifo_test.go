package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallCopiesAndCreatesParents(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "app")
	mustWriteFile(t, src, "binary content")
	dst := filepath.Join(dir, "bin", "app")

	code, _, errOut := run(t, "install", "-D", "-m", "755", src, dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "binary content" {
		t.Errorf("installed content = %q, want %q", got, "binary content")
	}
}

func TestInstallMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "install", "onlyone")
	if code == 0 {
		t.Error("expected nonzero exit for missing DEST argument")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestMkfifoReportsUnsupported(t *testing.T) {
	code, _, errOut := run(t, "mkfifo", "queue")
	if code == 0 {
		t.Error("expected nonzero exit since mkfifo is unsupported")
	}
	if errOut == "" {
		t.Error("expected an explanatory error message")
	}
}
