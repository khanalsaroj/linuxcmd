package commands

import (
	"strings"
	"testing"
)

func TestNetstatRuns(t *testing.T) {
	code, out, errOut := run(t, "netstat")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty netstat output")
	}
}

func TestArpRuns(t *testing.T) {
	code, _, errOut := run(t, "arp")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}

func TestRouteRuns(t *testing.T) {
	code, out, errOut := run(t, "route")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty route output")
	}
}

func TestSsRuns(t *testing.T) {
	code, _, errOut := run(t, "ss", "-tln")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}
