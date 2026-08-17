package commands

import (
	"strings"
	"testing"
)

func TestUnameDefault(t *testing.T) {
	code, out, errOut := run(t, "uname")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "Windows_NT" {
		t.Errorf("uname output = %q, want %q", strings.TrimSpace(out), "Windows_NT")
	}
}

func TestUnameAll(t *testing.T) {
	code, out, errOut := run(t, "uname", "-a")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Windows_NT") {
		t.Errorf("uname -a output = %q, want it to contain Windows_NT", out)
	}
}

func TestArchPrintsArchitecture(t *testing.T) {
	code, out, errOut := run(t, "arch")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty arch output")
	}
}

func TestNprocPrintsPositiveNumber(t *testing.T) {
	code, out, errOut := run(t, "nproc")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "0" || strings.TrimSpace(out) == "" {
		t.Errorf("expected a positive processor count, got %q", out)
	}
}
