package commands

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestShufPreservesAllLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\nd\ne\n")

	code, out, errOut := run(t, "shuf", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	want := []string{"a", "b", "c", "d", "e"}
	sort.Strings(got)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("shuf output missing/altered lines: got %v, want set %v", got, want)
			break
		}
	}
}

func TestShufWithCount(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nb\nc\nd\ne\n")

	code, out, errOut := run(t, "shuf", "-n", "2", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(got) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(got), got)
	}
}
