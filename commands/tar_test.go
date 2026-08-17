package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTarCreateAndExtractPlain(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "hello.txt")
	mustWriteFile(t, f, "hello world")

	archive := filepath.Join(dir, "out.tar")
	code, _, errOut := run(t, "tar", "-c", "-f", archive, f)
	if code != 0 {
		t.Fatalf("tar -c exit code = %d, stderr = %q", code, errOut)
	}

	extractDir := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = runIn(t, extractDir, "tar", "-x", "-f", archive)
	if code != 0 {
		t.Fatalf("tar -x exit code = %d, stderr = %q", code, errOut)
	}

	relPath, err := filepath.Rel(filepath.Dir(f), f)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, relPath))
	if err != nil {
		t.Fatalf("expected extracted file, got error: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("extracted content = %q, want %q", got, "hello world")
	}
}

func TestTarCreateAndExtractGzip(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "hello.txt")
	mustWriteFile(t, f, "compressed hello")

	archive := filepath.Join(dir, "out.tgz")
	code, _, errOut := run(t, "tar", "-czf", archive, f)
	if code != 0 {
		t.Fatalf("tar -czf exit code = %d, stderr = %q", code, errOut)
	}

	extractDir := filepath.Join(dir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut = runIn(t, extractDir, "tar", "-xzf", archive)
	if code != 0 {
		t.Fatalf("tar -xzf exit code = %d, stderr = %q", code, errOut)
	}

	relPath, err := filepath.Rel(filepath.Dir(f), f)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, relPath))
	if err != nil {
		t.Fatalf("expected extracted file, got error: %v", err)
	}
	if string(got) != "compressed hello" {
		t.Errorf("extracted content = %q, want %q", got, "compressed hello")
	}
}

func TestTarListMode(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "hello.txt")
	mustWriteFile(t, f, "x")

	archive := filepath.Join(dir, "out.tar")
	code, _, errOut := run(t, "tar", "-cf", archive, f)
	if code != 0 {
		t.Fatalf("tar -cf exit code = %d, stderr = %q", code, errOut)
	}

	code, out, errOut := run(t, "tar", "-tf", archive)
	if code != 0 {
		t.Fatalf("tar -tf exit code = %d, stderr = %q", code, errOut)
	}
	if out == "" {
		t.Error("expected tar -t to list entries")
	}
}

func TestTarMissingArchiveFlag(t *testing.T) {
	code, _, errOut := run(t, "tar", "-c")
	if code == 0 {
		t.Error("expected nonzero exit when -f is missing")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
