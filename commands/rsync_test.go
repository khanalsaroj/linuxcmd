package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRsyncCopiesTree(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "alpha")
	mustWriteFile(t, filepath.Join(src, "sub", "b.txt"), "beta")

	code, _, errOut := run(t, "rsync", "-a", src+string(filepath.Separator), dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, filepath.Join(dst, "a.txt")); got != "alpha" {
		t.Errorf("copied file = %q, want %q", got, "alpha")
	}
	if got := mustReadFile(t, filepath.Join(dst, "sub", "b.txt")); got != "beta" {
		t.Errorf("copied nested file = %q, want %q", got, "beta")
	}
}

// The trailing slash on the source is rsync's most consequential piece
// of syntax: "src/" copies the contents, "src" copies the directory.
func TestRsyncTrailingSlashSelectsContents(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "x")

	withSlash := filepath.Join(root, "with")
	if code, _, errOut := run(t, "rsync", "-a", src+"/", withSlash); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(withSlash, "a.txt")); err != nil {
		t.Errorf("with a trailing slash the contents should land directly in the destination: %v", err)
	}

	withoutSlash := filepath.Join(root, "without")
	if code, _, errOut := run(t, "rsync", "-a", src, withoutSlash); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(withoutSlash, "src", "a.txt")); err != nil {
		t.Errorf("without a trailing slash the directory itself should be copied: %v", err)
	}
}

// A second run must transfer nothing, which is the whole point of the
// size-and-time quick check.
func TestRsyncSkipsUnchangedFiles(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "alpha")

	if code, _, errOut := run(t, "rsync", "-av", src+"/", dst); code != 0 {
		t.Fatalf("first run: exit code = %d, stderr = %q", code, errOut)
	}
	code, out, errOut := run(t, "rsync", "-av", src+"/", dst)
	if code != 0 {
		t.Fatalf("second run: exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "0 transferred") {
		t.Errorf("second run transferred files it should have skipped:\n%s", out)
	}
}

func TestRsyncCopiesChangedFile(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	srcFile := filepath.Join(src, "a.txt")
	mustWriteFile(t, srcFile, "first")

	if code, _, errOut := run(t, "rsync", "-a", src+"/", dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	// A different length is enough for the quick check to notice.
	mustWriteFile(t, srcFile, "second version")
	if code, _, errOut := run(t, "rsync", "-a", src+"/", dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, filepath.Join(dst, "a.txt")); got != "second version" {
		t.Errorf("destination = %q, want the updated content", got)
	}
}

func TestRsyncDryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "alpha")

	code, out, errOut := run(t, "rsync", "-avn", src+"/", dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "dry run") {
		t.Errorf("dry-run output = %q, want it labelled as a dry run", out)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err == nil {
		t.Error("a dry run must not create files")
	}
}

func TestRsyncDeleteRemovesExtraneousFiles(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "keep.txt"), "keep")
	mustWriteFile(t, filepath.Join(dst, "stale.txt"), "stale")

	code, _, errOut := run(t, "rsync", "-a", "--delete", src+"/", dst)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dst, "stale.txt")); err == nil {
		t.Error("--delete should have removed the extraneous file")
	}
	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Errorf("--delete removed a file that exists in the source: %v", err)
	}
}

func TestRsyncWithoutDeleteKeepsExtraneousFiles(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "keep.txt"), "keep")
	mustWriteFile(t, filepath.Join(dst, "extra.txt"), "extra")

	if code, _, errOut := run(t, "rsync", "-a", src+"/", dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dst, "extra.txt")); err != nil {
		t.Error("without --delete an extraneous file must be left alone")
	}
}

func TestRsyncExclude(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "keep.txt"), "keep")
	mustWriteFile(t, filepath.Join(src, "skip.log"), "skip")

	if code, _, errOut := run(t, "rsync", "-a", "--exclude", "*.log", src+"/", dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(dst, "skip.log")); err == nil {
		t.Error("--exclude did not skip the matching file")
	}
	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Errorf("--exclude skipped a file it should not have: %v", err)
	}
}

func TestRsyncUpdateKeepsNewerDestination(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "old")
	mustWriteFile(t, filepath.Join(dst, "a.txt"), "newer content")

	future := time.Now().Add(1 * time.Hour)
	if err := os.Chtimes(filepath.Join(dst, "a.txt"), future, future); err != nil {
		t.Fatal(err)
	}

	if code, _, errOut := run(t, "rsync", "-a", "-u", src+"/", dst); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if got := mustReadFile(t, filepath.Join(dst, "a.txt")); got != "newer content" {
		t.Errorf("-u overwrote a newer destination: %q", got)
	}
}

func TestRsyncRefusesDirectoryWithoutRecursive(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "x")

	_, _, errOut := run(t, "rsync", src, filepath.Join(root, "dst"))
	if !strings.Contains(errOut, "skipping directory") {
		t.Errorf("stderr = %q, want rsync's skipping-directory notice", errOut)
	}
}

func TestRsyncRejectsRemoteSpecs(t *testing.T) {
	dir := t.TempDir()
	for _, spec := range []string{"user@host:/data", "host:/data", "rsync://host/mod", "host::module"} {
		code, _, errOut := run(t, "rsync", "-a", spec, dir)
		if code == 0 {
			t.Errorf("%s: expected a nonzero exit for a remote spec", spec)
		}
		if !strings.Contains(errOut, "remote transfers are not supported") {
			t.Errorf("%s: stderr = %q, want a clear refusal", spec, errOut)
		}
	}
}

// A Windows drive letter must never be mistaken for a hostname, or every
// absolute path on Windows would be rejected as remote.
func TestIsRemoteSpec(t *testing.T) {
	tests := []struct {
		spec string
		want bool
	}{
		{`C:\Users\x`, false},
		{"C:/Users/x", false},
		{"d:/data", false},
		{"./relative", false},
		{"/tmp/thing", false},
		{"plain.txt", false},
		{`dir\sub:weird`, false},
		{"host:/data", true},
		{"user@host:/data", true},
		{"rsync://host/module", true},
		{"host::module", true},
	}
	for _, tt := range tests {
		if got := isRemoteSpec(tt.spec); got != tt.want {
			t.Errorf("isRemoteSpec(%q) = %v, want %v", tt.spec, got, tt.want)
		}
	}
}

func TestRsyncRequiresTwoOperands(t *testing.T) {
	code, _, errOut := run(t, "rsync", "-a", "only-one")
	if code == 0 {
		t.Error("expected a nonzero exit with a single operand")
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want a usage line", errOut)
	}
}

func TestMatchesExclude(t *testing.T) {
	tests := []struct {
		rel     string
		pattern string
		want    bool
	}{
		{"a.log", "*.log", true},
		{filepath.Join("sub", "a.log"), "*.log", true},
		{filepath.Join("node_modules", "x.js"), "node_modules", true},
		{"a.txt", "*.log", false},
	}
	for _, tt := range tests {
		if got := matchesExclude(tt.rel, tt.pattern); got != tt.want {
			t.Errorf("matchesExclude(%q, %q) = %v, want %v", tt.rel, tt.pattern, got, tt.want)
		}
	}
}
