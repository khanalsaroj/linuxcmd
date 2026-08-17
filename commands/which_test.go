package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWhichFindsExecutableOnPath(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "foocmd.exe")
	mustWriteFile(t, exe, "")

	origPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Setenv("PATH", origPath) })

	code, out, errOut := run(t, "which", "foocmd")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out == "" {
		t.Errorf("expected which to print a path, got empty output")
	}
}

func TestWhichNotFound(t *testing.T) {
	code, _, _ := run(t, "which", "definitely-not-a-real-command-xyz")
	if code == 0 {
		t.Error("expected nonzero exit for a command not on PATH")
	}
}
