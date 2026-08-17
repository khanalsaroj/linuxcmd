package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGrepMatchesLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello world\ngoodbye\nhello again\n")

	code, out, errOut := run(t, "grep", "hello", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "hello world") || !strings.Contains(out, "hello again") {
		t.Errorf("expected both hello lines, got %q", out)
	}
	if strings.Contains(out, "goodbye") {
		t.Errorf("did not expect 'goodbye' line in output: %q", out)
	}
}

func TestGrepNoMatchExitCode(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "nothing relevant here\n")

	code, _, _ := run(t, "grep", "zzz-not-present", f)
	if code != 1 {
		t.Errorf("exit code = %d, want 1 for no matches", code)
	}
}

func TestGrepMissingFileExitCode(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "grep", "pattern", filepath.Join(dir, "nope.txt"))
	if code != 2 {
		t.Errorf("exit code = %d, want 2 for a file error", code)
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestGrepCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "HELLO\n")

	code, out, errOut := run(t, "grep", "-i", "hello", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "HELLO") {
		t.Errorf("expected case-insensitive match, got %q", out)
	}
}

func TestGrepInvertMatch(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "keep\nremove\nkeep2\n")

	code, out, errOut := run(t, "grep", "-v", "remove", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "remove") {
		t.Errorf("did not expect 'remove' line with -v, got %q", out)
	}
	if !strings.Contains(out, "keep") || !strings.Contains(out, "keep2") {
		t.Errorf("expected kept lines, got %q", out)
	}
}

func TestGrepLineNumbers(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "a\nmatch\nb\n")

	code, out, errOut := run(t, "grep", "-n", "match", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "2:match") {
		t.Errorf("expected line-numbered output '2:match', got %q", out)
	}
}

func TestGrepUsageError(t *testing.T) {
	code, _, errOut := run(t, "grep")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 for usage error", code)
	}
	if errOut == "" {
		t.Error("expected a usage message")
	}
}

func TestFindByName(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), "x")
	mustWriteFile(t, filepath.Join(dir, "b.log"), "x")
	mustWriteFile(t, filepath.Join(dir, "sub", "c.txt"), "x")

	code, out, errOut := run(t, "find", dir, "-name", "*.txt")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "c.txt") {
		t.Errorf("expected both .txt files found, got %q", out)
	}
	if strings.Contains(out, "b.log") {
		t.Errorf("did not expect b.log in -name *.txt results: %q", out)
	}
}

func TestFindByType(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "file.txt"), "x")

	code, out, errOut := run(t, "find", dir, "-type", "d")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "file.txt") {
		t.Errorf("-type d should not list files, got %q", out)
	}
	if !strings.Contains(out, dir) {
		t.Errorf("expected the root directory itself in -type d results, got %q", out)
	}
}

func TestFindMissingPath(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "find", filepath.Join(dir, "nope"))
	if code == 0 {
		t.Error("expected nonzero exit for a missing start path")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestFindUnknownPredicate(t *testing.T) {
	code, _, errOut := run(t, "find", "-bogus")
	if code == 0 {
		t.Error("expected nonzero exit for an unknown predicate")
	}
	if !strings.Contains(errOut, "unknown predicate") {
		t.Errorf("expected 'unknown predicate' error, got %q", errOut)
	}
}
