package commands

import (
	"strings"
	"testing"
)

func TestEnvListsVariables(t *testing.T) {
	code, out, errOut := run(t, "env")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty environment listing")
	}
}

func TestEnvRunsCommandWithOverride(t *testing.T) {
	code, out, errOut := run(t, "env", "GREETING=hello", "cmd", "/c", "echo", "%GREETING%")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("env output = %q, want it to contain the overridden value", out)
	}
}

func TestPrintenvSpecificVariable(t *testing.T) {
	t.Setenv("LINUXCMD_TEST_VAR", "test-value")
	code, out, errOut := run(t, "printenv", "LINUXCMD_TEST_VAR")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "test-value" {
		t.Errorf("printenv output = %q, want %q", strings.TrimSpace(out), "test-value")
	}
}

func TestPrintenvUnknownVariable(t *testing.T) {
	code, _, _ := run(t, "printenv", "DEFINITELY_NOT_SET_XYZ")
	if code == 0 {
		t.Error("expected nonzero exit for an unset variable")
	}
}
