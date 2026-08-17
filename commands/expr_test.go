package commands

import (
	"strings"
	"testing"
)

func TestExprArithmetic(t *testing.T) {
	code, out, errOut := run(t, "expr", "3", "+", "4")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "7" {
		t.Errorf("expr 3 + 4 output = %q, want %q", strings.TrimSpace(out), "7")
	}
}

func TestExprStringEquality(t *testing.T) {
	code, out, errOut := run(t, "expr", "abc", "=", "abc")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "1" {
		t.Errorf("expr abc = abc output = %q, want %q", strings.TrimSpace(out), "1")
	}
}

func TestExprComparisonFalse(t *testing.T) {
	code, out, _ := run(t, "expr", "3", ">", "5")
	if code == 0 {
		t.Error("expected nonzero exit for a false comparison")
	}
	if strings.TrimSpace(out) != "0" {
		t.Errorf("expr 3 > 5 output = %q, want %q", strings.TrimSpace(out), "0")
	}
}

func TestExprDivisionByZero(t *testing.T) {
	code, _, errOut := run(t, "expr", "3", "/", "0")
	if code == 0 {
		t.Error("expected nonzero exit for division by zero")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
