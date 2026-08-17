package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadlinkPrintsTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	mustWriteFile(t, target, "x")
	link := filepath.Join(dir, "link.txt")

	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation not permitted in this environment: %v", err)
	}

	code, out, errOut := run(t, "readlink", link)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != target {
		t.Errorf("readlink output = %q, want %q", strings.TrimSpace(out), target)
	}
}

func TestReadlinkNotASymlink(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "readlink", f)
	if code == 0 {
		t.Error("expected nonzero exit for a non-symlink target")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
