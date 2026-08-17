package commands

import (
	"strings"
	"testing"
)

func TestTtyNotATtyForNonFileStdin(t *testing.T) {
	// The test harness's run() helper feeds stdin from a bytes.Reader,
	// not a real console, so tty should report "not a tty".
	code, out, errOut := run(t, "tty")
	if code == 0 {
		t.Error("expected nonzero exit for a non-console stdin")
	}
	if strings.TrimSpace(out) != "not a tty" {
		t.Errorf("tty output = %q, want %q", strings.TrimSpace(out), "not a tty")
	}
	_ = errOut
}

func TestHostidPrintsHexIdentifier(t *testing.T) {
	code, out, errOut := run(t, "hostid")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	id := strings.TrimSpace(out)
	if len(id) != 8 {
		t.Errorf("hostid output = %q, want an 8-character hex string", id)
	}
}
