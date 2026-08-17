package commands

import (
	"strings"
	"testing"
)

func TestSyncSucceeds(t *testing.T) {
	code, _, _ := run(t, "sync")
	if code != 0 {
		t.Errorf("expected sync to exit 0, got %d", code)
	}
}

func TestUmaskReportsDefault(t *testing.T) {
	code, out, errOut := run(t, "umask")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected umask to print a mask")
	}
}

func TestUmaskInvalidValue(t *testing.T) {
	code, _, errOut := run(t, "umask", "999")
	if code == 0 {
		t.Error("expected nonzero exit for an invalid octal mask")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
