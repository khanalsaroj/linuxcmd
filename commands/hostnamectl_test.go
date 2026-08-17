package commands

import (
	"strings"
	"testing"
)

func TestHostnamectlStatus(t *testing.T) {
	code, out, errOut := run(t, "hostnamectl")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Static hostname:") {
		t.Errorf("hostnamectl output = %q, want a hostname line", out)
	}
}

func TestHostnamectlUnsupportedVerb(t *testing.T) {
	code, _, errOut := run(t, "hostnamectl", "set-hostname", "new-name")
	if code == 0 {
		t.Error("expected nonzero exit for unsupported verb")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
