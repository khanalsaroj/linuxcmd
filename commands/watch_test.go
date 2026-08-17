package commands

import (
	"strings"
	"testing"
)

func TestWatchBoundedIterations(t *testing.T) {
	code, out, errOut := run(t, "watch", "-n", "0.05", "-c", "2", "cmd", "/c", "echo", "hi")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Count(out, "hi") != 2 {
		t.Errorf("expected 2 iterations of output, got %q", out)
	}
}

func TestWatchMissingCommand(t *testing.T) {
	code, _, errOut := run(t, "watch")
	if code == 0 {
		t.Error("expected nonzero exit for missing command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
