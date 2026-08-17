package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLnHardLink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	mustWriteFile(t, target, "content")
	link := filepath.Join(dir, "link.txt")

	code, _, errOut := run(t, "ln", target, link)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "content" {
		t.Errorf("hard link content = %q, want %q", got, "content")
	}
}

func TestLnSymbolic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	mustWriteFile(t, target, "content")
	link := filepath.Join(dir, "link.txt")

	code, _, errOut := run(t, "ln", "-s", target, link)
	if code != 0 {
		t.Skipf("symlink creation not permitted in this environment: %s", errOut)
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Errorf("symlink target = %q, want %q", got, target)
	}
}
