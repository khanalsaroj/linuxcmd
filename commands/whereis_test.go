package commands

import (
	"strings"
	"testing"
)

func TestWhereisFindsBuiltin(t *testing.T) {
	code, out, errOut := run(t, "whereis", "cat")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "linuxcmd built-in") {
		t.Errorf("whereis output = %q, want it to mark 'cat' as a linuxcmd built-in", out)
	}
}

func TestWhereisMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "whereis")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
