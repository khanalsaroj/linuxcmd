package commands

import (
	"strings"
	"testing"
)

func TestTopSingleIteration(t *testing.T) {
	code, out, errOut := run(t, "top", "-n", "1")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "PID") || !strings.Contains(out, "CMD") {
		t.Errorf("top output = %q, want a PID/CMD header", out)
	}
}

func TestTopInvalidIterationCount(t *testing.T) {
	code, _, errOut := run(t, "top", "-n", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for invalid -n value")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
