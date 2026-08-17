package commands

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestSHA256SumPrintsHash(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello")

	code, out, errOut := run(t, "sha256sum", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := fmt.Sprintf("%x", sha256.Sum256([]byte("hello")))
	if !strings.HasPrefix(out, want) {
		t.Errorf("sha256sum output = %q, want prefix %q", out, want)
	}
	if !strings.Contains(out, f) {
		t.Errorf("expected output to include filename %q, got %q", f, out)
	}
}

func TestSHA256SumCheck(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello")

	_, sumOut, errOut := run(t, "sha256sum", f)
	if errOut != "" {
		t.Fatalf("unexpected stderr computing checksum: %q", errOut)
	}
	sumFile := filepath.Join(dir, "checksums.sha256")
	mustWriteFile(t, sumFile, sumOut)

	code, out, errOut := run(t, "sha256sum", "-c", sumFile)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in check output, got %q", out)
	}
}

func TestSHA256SumCheckDetectsMismatch(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "hello")

	sumFile := filepath.Join(dir, "checksums.sha256")
	mustWriteFile(t, sumFile, "0000000000000000000000000000000000000000000000000000000000000000  "+f+"\n")

	code, out, _ := run(t, "sha256sum", "-c", sumFile)
	if code == 0 {
		t.Error("expected nonzero exit for a mismatched checksum")
	}
	if !strings.Contains(out, "FAILED") {
		t.Errorf("expected FAILED in check output, got %q", out)
	}
}

func TestMD5SumStdin(t *testing.T) {
	code, out, errOut := runWithStdin(t, "md5sum", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "-") {
		t.Errorf("expected stdin marker '-' in output, got %q", out)
	}
}

func TestCksumPrintsCountAndCRC(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "abc")

	code, out, errOut := run(t, "cksum", f)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	fields := strings.Fields(out)
	if len(fields) != 3 || fields[1] != "3" || fields[2] != f {
		t.Errorf("cksum output = %q", out)
	}
}
