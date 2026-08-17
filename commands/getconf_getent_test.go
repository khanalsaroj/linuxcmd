package commands

import (
	"strings"
	"testing"
)

func TestGetconfKnownValue(t *testing.T) {
	code, out, errOut := run(t, "getconf", "PAGE_SIZE")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "4096" {
		t.Errorf("getconf PAGE_SIZE output = %q, want %q", strings.TrimSpace(out), "4096")
	}
}

func TestGetconfUnknownValue(t *testing.T) {
	code, _, errOut := run(t, "getconf", "NOT_A_REAL_VAR")
	if code == 0 {
		t.Error("expected nonzero exit for an unknown variable")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestGetentPasswdCurrentUser(t *testing.T) {
	code, out, errOut := run(t, "getent", "passwd")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, ":x:") {
		t.Errorf("getent passwd output = %q, want passwd-style fields", out)
	}
}

func TestGetentUnknownDatabase(t *testing.T) {
	code, _, errOut := run(t, "getent", "bogus")
	if code == 0 {
		t.Error("expected nonzero exit for an unknown database")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
