package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCsplitByRegex(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	mustWriteFile(t, src, "intro\nSECTION A\na1\na2\nSECTION B\nb1\n")

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runIn(t, outDir, "csplit", src, "/^SECTION/", "/^SECTION/")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out == "" {
		t.Error("expected csplit to print created filenames")
	}

	first, err := os.ReadFile(filepath.Join(outDir, "xx00"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != "intro\n" {
		t.Errorf("xx00 content = %q, want %q", first, "intro\n")
	}
}

func TestCsplitByLineNumber(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	mustWriteFile(t, src, "a\nb\nc\nd\n")

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := runIn(t, outDir, "csplit", src, "3")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	first, err := os.ReadFile(filepath.Join(outDir, "xx00"))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != "a\nb\n" {
		t.Errorf("xx00 content = %q, want %q", first, "a\nb\n")
	}
}
