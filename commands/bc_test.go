package commands

import (
	"strings"
	"testing"
)

func TestBcSimpleExpression(t *testing.T) {
	code, out, errOut := runWithStdin(t, "bc", "3 + 4 * 2\n")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "11" {
		t.Errorf("bc output = %q, want %q", strings.TrimSpace(out), "11")
	}
}

func TestBcParentheses(t *testing.T) {
	code, out, errOut := runWithStdin(t, "bc", "(3 + 4) * 2\n")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "14" {
		t.Errorf("bc output = %q, want %q", strings.TrimSpace(out), "14")
	}
}

func TestBcDivisionByZero(t *testing.T) {
	code, _, errOut := runWithStdin(t, "bc", "1 / 0\n")
	if code == 0 {
		t.Error("expected nonzero exit for division by zero")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestBcArgumentMode(t *testing.T) {
	code, out, errOut := run(t, "bc", "2 ^ 10")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "1024" {
		t.Errorf("bc 2 ^ 10 output = %q, want %q", strings.TrimSpace(out), "1024")
	}
}
