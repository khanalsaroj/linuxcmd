package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNlNumbersAllLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\n\nb\n")

	code, out, errOut := run(t, "nl", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 output lines, got %q", out)
	}
	if !strings.Contains(lines[0], "1") || !strings.Contains(lines[0], "a") {
		t.Errorf("expected first line numbered, got %q", lines[0])
	}
}
