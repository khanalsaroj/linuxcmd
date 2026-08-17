package commands

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func selfPattern(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("could not determine own executable: %v", err)
	}
	base := filepath.Base(exe)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func TestPgrepFindsRunningProcess(t *testing.T) {
	pattern := selfPattern(t)

	code, out, errOut := run(t, "pgrep", pattern)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, strconv.Itoa(os.Getpid())) {
		t.Errorf("expected pgrep to list our own PID %d, got %q", os.Getpid(), out)
	}
}

func TestPgrepNoMatch(t *testing.T) {
	code, _, _ := run(t, "pgrep", "definitely-not-a-real-process-xyz")
	if code == 0 {
		t.Error("expected nonzero exit for no matches")
	}
}

func TestPgrepListName(t *testing.T) {
	pattern := selfPattern(t)

	code, out, errOut := run(t, "pgrep", "-l", pattern)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(strings.ToLower(out), strings.ToLower(pattern)) {
		t.Errorf("expected -l output to include the process name, got %q", out)
	}
}
