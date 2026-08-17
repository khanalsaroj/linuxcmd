package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCPFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	mustWriteFile(t, src, "hello")

	code, _, errOut := run(t, "cp", src, dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hello" {
		t.Errorf("expected dst.txt to contain 'hello', got %q (err=%v)", data, err)
	}
}

func TestCPMissingSource(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "cp", filepath.Join(dir, "nope.txt"), filepath.Join(dir, "dst.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing source")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestCPDirectoryWithoutDashRFails(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "srcdir")
	if err := os.Mkdir(src, 0755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "cp", src, filepath.Join(dir, "dstdir"))
	if code == 0 {
		t.Error("expected nonzero exit copying a directory without -r")
	}
	if !strings.Contains(errOut, "-r not specified") {
		t.Errorf("expected -r hint in error, got %q", errOut)
	}
}

func TestCPRecursive(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "srcdir")
	mustWriteFile(t, filepath.Join(src, "nested", "f.txt"), "content")
	dst := filepath.Join(dir, "dstdir")

	code, _, errOut := run(t, "cp", "-r", src, dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	data, err := os.ReadFile(filepath.Join(dst, "nested", "f.txt"))
	if err != nil || string(data) != "content" {
		t.Errorf("expected recursive copy of nested/f.txt, got err=%v data=%q", err, data)
	}
}

func TestCPMultipleSourcesRequireDirectoryDest(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "a")
	mustWriteFile(t, b, "b")
	notADir := filepath.Join(dir, "notadir.txt")
	mustWriteFile(t, notADir, "x")

	code, _, errOut := run(t, "cp", a, b, notADir)
	if code == 0 {
		t.Error("expected nonzero exit when copying multiple sources to a non-directory")
	}
	if !strings.Contains(errOut, "is not a directory") {
		t.Errorf("expected 'is not a directory' error, got %q", errOut)
	}
}

func TestMVRename(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "old.txt")
	dst := filepath.Join(dir, "new.txt")
	mustWriteFile(t, src, "data")

	code, _, errOut := run(t, "mv", src, dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("expected source to no longer exist after mv")
	}
	if data, err := os.ReadFile(dst); err != nil || string(data) != "data" {
		t.Errorf("expected new.txt to contain 'data', got err=%v data=%q", err, data)
	}
}

func TestMVMissingSource(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "mv", filepath.Join(dir, "nope.txt"), filepath.Join(dir, "dst.txt"))
	if code == 0 {
		t.Error("expected nonzero exit for missing source")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestRMFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "rm", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}
}

func TestRMDirectoryWithoutDashRFails(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}

	code, _, errOut := run(t, "rm", sub)
	if code == 0 {
		t.Error("expected nonzero exit removing a directory without -r")
	}
	if !strings.Contains(errOut, "Is a directory") {
		t.Errorf("expected 'Is a directory' error, got %q", errOut)
	}
}

func TestRMRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	mustWriteFile(t, filepath.Join(sub, "f.txt"), "x")

	code, _, errOut := run(t, "rm", "-r", sub)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(sub); !os.IsNotExist(err) {
		t.Error("expected directory tree to be removed")
	}
}

func TestRMForceIgnoresMissing(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := run(t, "rm", "-f", filepath.Join(dir, "nope.txt"))
	if code != 0 {
		t.Errorf("expected exit 0 with -f on missing file, got %d, stderr=%q", code, errOut)
	}
}

func TestRMMissingOperandExitCode(t *testing.T) {
	code, _, errOut := run(t, "rm")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("expected 'missing operand' error, got %q", errOut)
	}
}

func TestRMGlobExpansion(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.txt"), "x")
	mustWriteFile(t, filepath.Join(dir, "b.txt"), "x")
	mustWriteFile(t, filepath.Join(dir, "keep.log"), "x")

	code, _, errOut := run(t, "rm", filepath.Join(dir, "*.txt"))
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.txt")); !os.IsNotExist(err) {
		t.Error("expected a.txt removed by glob")
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.log")); err != nil {
		t.Error("expected keep.log to survive a *.txt glob")
	}
}
