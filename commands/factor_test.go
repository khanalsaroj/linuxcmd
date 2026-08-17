package commands

import (
	"strings"
	"testing"
)

func TestFactorComposite(t *testing.T) {
	code, out, errOut := run(t, "factor", "360")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "360: 2 2 2 3 3 5"
	if strings.TrimSpace(out) != want {
		t.Errorf("factor 360 output = %q, want %q", strings.TrimSpace(out), want)
	}
}

func TestFactorPrime(t *testing.T) {
	code, out, errOut := run(t, "factor", "17")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "17: 17" {
		t.Errorf("factor 17 output = %q, want %q", strings.TrimSpace(out), "17: 17")
	}
}

func TestFactorInvalidInput(t *testing.T) {
	code, _, errOut := run(t, "factor", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for invalid input")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
