package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffIdenticalFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "same\ncontent\n")
	mustWriteFile(t, b, "same\ncontent\n")

	code, out, errOut := run(t, "diff", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "" {
		t.Errorf("expected no output for identical files, got %q", out)
	}
}

func TestDiffNormalFormat(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "one\ntwo\nthree\n")
	mustWriteFile(t, b, "one\ntwo-changed\nthree\n")

	code, out, _ := run(t, "diff", a, b)
	if code == 0 {
		t.Error("expected nonzero exit for differing files")
	}
	if !strings.Contains(out, "< two\n") || !strings.Contains(out, "> two-changed\n") {
		t.Errorf("diff output = %q, want lines marked with < and >", out)
	}
}

func TestDiffUnifiedFormat(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "one\ntwo\n")
	mustWriteFile(t, b, "one\ntwo-changed\n")

	code, out, _ := run(t, "diff", "-u", a, b)
	if code == 0 {
		t.Error("expected nonzero exit for differing files")
	}
	if !strings.Contains(out, "--- "+a) || !strings.Contains(out, "+++ "+b) {
		t.Errorf("expected unified headers, got %q", out)
	}
	if !strings.Contains(out, "-two\n") || !strings.Contains(out, "+two-changed\n") {
		t.Errorf("expected +/- marked lines, got %q", out)
	}
}
