package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMkdir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "newdir")

	code, _, errOut := run(t, "mkdir", target)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		t.Errorf("expected %q to be created as a directory", target)
	}
}

func TestMkdirAlreadyExists(t *testing.T) {
	dir := t.TempDir()

	code, _, errOut := run(t, "mkdir", dir)
	if code == 0 {
		t.Error("expected nonzero exit when directory already exists")
	}
	if !strings.Contains(errOut, "File exists") {
		t.Errorf("expected 'File exists' error, got %q", errOut)
	}
}

func TestMkdirParentsFlag(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")

	code, _, errOut := run(t, "mkdir", "-p", target)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if info, err := os.Stat(target); err != nil || !info.IsDir() {
		t.Errorf("expected nested directory %q to exist", target)
	}
}

func TestMkdirMissingParentWithoutDashP(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "missing", "child")

	code, _, errOut := run(t, "mkdir", target)
	if code == 0 {
		t.Error("expected nonzero exit without -p when parent is missing")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestMkdirMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "mkdir")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("expected 'missing operand' error, got %q", errOut)
	}
}

func TestRmdirEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "empty")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "rmdir", target)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected %q to be removed", target)
	}
}

func TestRmdirNonEmptyDirectoryFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nonempty")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(target, "f.txt"), "x")

	code, _, errOut := run(t, "rmdir", target)
	if code == 0 {
		t.Error("expected nonzero exit removing a non-empty directory")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestRmdirParentsFlag(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "rmdir", "-p", target)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "a")); !os.IsNotExist(err) {
		t.Error("expected -p to remove now-empty parent directories too")
	}
}
