package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZipAndUnzipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "hello.txt")
	mustWriteFile(t, f, "hello world")

	archive := filepath.Join(dir, "out.zip")
	code, _, errOut := run(t, "zip", archive, f)
	if code != 0 {
		t.Fatalf("zip exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("expected archive to be created: %v", err)
	}

	extractDir := filepath.Join(dir, "extracted")
	code, _, errOut = run(t, "unzip", archive, "-d", extractDir)
	if code != 0 {
		t.Fatalf("unzip exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Errorf("extracted content = %q, want %q", got, "hello world")
	}
}

func TestZipRecursiveDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "src", "nested")
	f := filepath.Join(sub, "a.txt")
	mustWriteFile(t, f, "AAA")

	archive := filepath.Join(dir, "out.zip")
	code, _, errOut := run(t, "zip", "-r", archive, filepath.Join(dir, "src"))
	if code != 0 {
		t.Fatalf("zip -r exit code = %d, stderr = %q", code, errOut)
	}

	extractDir := filepath.Join(dir, "extracted")
	code, _, errOut = run(t, "unzip", archive, "-d", extractDir)
	if code != 0 {
		t.Fatalf("unzip exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "src", "nested", "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "AAA" {
		t.Errorf("extracted content = %q, want %q", got, "AAA")
	}
}

func TestZipDirectoryWithoutRecursiveFails(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "src")
	mustWriteFile(t, filepath.Join(sub, "a.txt"), "AAA")

	archive := filepath.Join(dir, "out.zip")
	code, _, errOut := run(t, "zip", archive, sub)
	if code == 0 {
		t.Error("expected nonzero exit when zipping a directory without -r")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
