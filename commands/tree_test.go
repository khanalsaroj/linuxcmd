package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestTreeListsEntries(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), "x")
	mustWriteFile(t, filepath.Join(dir, "sub", "b.txt"), "y")

	code, out, errOut := run(t, "tree", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "sub") || !strings.Contains(out, "b.txt") {
		t.Errorf("expected tree to list entries, got %q", out)
	}
	if !strings.Contains(out, "directories") || !strings.Contains(out, "files") {
		t.Errorf("expected summary line, got %q", out)
	}
}

func TestTreeDepthLimit(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), "x")
	mustWriteFile(t, filepath.Join(dir, "sub", "nested", "b.txt"), "y")

	code, out, errOut := run(t, "tree", "-L", "1", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "b.txt") {
		t.Errorf("expected -L 1 to exclude nested files, got %q", out)
	}
	if !strings.Contains(out, "sub") {
		t.Errorf("expected top-level 'sub' to still be listed, got %q", out)
	}
}

func TestTreeHidesDotFilesByDefault(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, ".hidden"), "x")
	mustWriteFile(t, filepath.Join(dir, "visible.txt"), "y")

	code, out, errOut := run(t, "tree", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, ".hidden") {
		t.Errorf("expected dotfile to be hidden by default, got %q", out)
	}
}
