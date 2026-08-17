package commands

import (
	"strings"
	"testing"
)

func TestTimeReportsElapsed(t *testing.T) {
	code, _, errOut := run(t, "time", "cmd", "/c", "exit", "0")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(errOut, "real") {
		t.Errorf("time stderr = %q, want it to report elapsed 'real' time", errOut)
	}
}

func TestTimeMissingCommand(t *testing.T) {
	code, _, errOut := run(t, "time")
	if code == 0 {
		t.Error("expected nonzero exit for missing command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
