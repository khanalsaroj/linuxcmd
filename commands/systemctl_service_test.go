package commands

import (
	"strings"
	"testing"
)

// Only the read-only "status" verb is exercised here (never
// start/stop/restart), since those change real system service state and
// require administrator rights.
func TestSystemctlStatus(t *testing.T) {
	code, out, errOut := run(t, "systemctl", "status", "Spooler")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty status output")
	}
}

func TestSystemctlUnknownVerb(t *testing.T) {
	code, _, errOut := run(t, "systemctl", "frobnicate", "Spooler")
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported verb")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestServiceStatus(t *testing.T) {
	code, out, errOut := run(t, "service", "Spooler", "status")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty status output")
	}
}
