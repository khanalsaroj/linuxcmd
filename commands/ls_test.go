package commands

import (
	"strings"
	"testing"
)

func TestLSListsFiles(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\a.txt`, "x")
	mustWriteFile(t, dir+`\b.txt`, "x")

	code, out, errOut := run(t, "ls", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "b.txt") {
		t.Errorf("output missing expected files: %q", out)
	}
}

func TestLSLongFormat(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\a.txt`, "hello")

	code, out, errOut := run(t, "ls", "-l", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	line := strings.TrimSpace(out)
	if !strings.HasPrefix(line, "-") {
		t.Errorf("long-format line should start with a file-type char, got %q", line)
	}
	if !strings.Contains(out, "a.txt") {
		t.Errorf("missing filename in long output: %q", out)
	}
	if !strings.Contains(out, "5") { // file size
		t.Errorf("missing file size in long output: %q", out)
	}
}

func TestLSCombinedFlags(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\visible.txt`, "x")

	code, out, errOut := run(t, "ls", "-la", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "visible.txt") {
		t.Errorf("expected visible.txt in -la output: %q", out)
	}
}

func TestLSHiddenFileRequiresDashA(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\.hidden`, "x")
	mustWriteFile(t, dir+`\shown.txt`, "x")

	_, out, _ := run(t, "ls", dir)
	if strings.Contains(out, ".hidden") {
		t.Errorf("ls without -a should not show dotfiles: %q", out)
	}

	_, out, _ = run(t, "ls", "-a", dir)
	if !strings.Contains(out, ".hidden") {
		t.Errorf("ls -a should show dotfiles: %q", out)
	}
}

func TestLSMissingPathExitCode(t *testing.T) {
	code, _, errOut := run(t, "ls", `Z:\does\not\exist\at\all`)
	if code == 0 {
		t.Error("expected nonzero exit code for missing path")
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("expected Linux-style error, got %q", errOut)
	}
}

func TestLSUnicodeFilename(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\héllo-世界.txt`, "x")

	code, out, errOut := run(t, "ls", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "héllo-世界.txt") {
		t.Errorf("expected unicode filename in output: %q", out)
	}
}

func TestLSFilenameWithSpaces(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\my file.txt`, "x")

	code, out, errOut := run(t, "ls", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "my file.txt") {
		t.Errorf("expected filename with spaces in output: %q", out)
	}
}

func TestLSEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := run(t, "ls", dir)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no output for empty directory, got %q", out)
	}
}

func TestLSRelativePath(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir+`\rel.txt`, "x")

	code, out, errOut := runIn(t, dir, "ls", ".")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "rel.txt") {
		t.Errorf("expected rel.txt via relative '.', got %q", out)
	}
}

func TestLSInvalidFlag(t *testing.T) {
	code, _, errOut := run(t, "ls", "-z")
	if code == 0 {
		t.Error("expected nonzero exit for invalid flag")
	}
	if errOut == "" {
		t.Error("expected an error message for invalid flag")
	}
}
