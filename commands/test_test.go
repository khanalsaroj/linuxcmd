package commands

import (
	"path/filepath"
	"testing"
)

func TestTestStringEquality(t *testing.T) {
	code, _, _ := run(t, "test", "abc", "=", "abc")
	if code != 0 {
		t.Errorf("expected equal strings to succeed, got exit %d", code)
	}
	code, _, _ = run(t, "test", "abc", "=", "xyz")
	if code == 0 {
		t.Errorf("expected unequal strings to fail")
	}
}

func TestTestIntegerComparison(t *testing.T) {
	code, _, _ := run(t, "test", "3", "-lt", "5")
	if code != 0 {
		t.Errorf("expected 3 -lt 5 to succeed, got exit %d", code)
	}
	code, _, _ = run(t, "test", "5", "-lt", "3")
	if code == 0 {
		t.Errorf("expected 5 -lt 3 to fail")
	}
}

func TestTestFileExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, _ := run(t, "test", "-f", f)
	if code != 0 {
		t.Errorf("expected -f on existing file to succeed, got exit %d", code)
	}

	code, _, _ = run(t, "test", "-d", f)
	if code == 0 {
		t.Errorf("expected -d on a regular file to fail")
	}

	code, _, _ = run(t, "test", "-d", dir)
	if code != 0 {
		t.Errorf("expected -d on a directory to succeed, got exit %d", code)
	}
}

func TestTestEmptyString(t *testing.T) {
	code, _, _ := run(t, "test", "-z", "")
	if code != 0 {
		t.Errorf("expected -z on empty string to succeed, got exit %d", code)
	}
	code, _, _ = run(t, "test", "-n", "abc")
	if code != 0 {
		t.Errorf("expected -n on non-empty string to succeed, got exit %d", code)
	}
}

func TestTestNegation(t *testing.T) {
	code, _, _ := run(t, "test", "!", "-f", filepath.Join(t.TempDir(), "nope.txt"))
	if code != 0 {
		t.Errorf("expected negated missing-file test to succeed, got exit %d", code)
	}
}
