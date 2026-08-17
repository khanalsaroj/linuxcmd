package commands

import (
	"strings"
	"testing"
)

func TestPstreePrintsProcessNames(t *testing.T) {
	pattern := selfPattern(t)

	code, out, errOut := run(t, "pstree")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected non-empty pstree output")
	}
	if !strings.Contains(strings.ToLower(out), strings.ToLower(pattern)) {
		t.Errorf("expected pstree output to include our own process name %q, got %q", pattern, out)
	}
}

func TestPstreeShowsPIDs(t *testing.T) {
	code, out, errOut := run(t, "pstree", "-p")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "(") {
		t.Errorf("expected -p output to include PIDs in parentheses, got %q", out)
	}
}
