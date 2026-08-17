package commands

import "testing"

func TestSttyNonConsoleStdin(t *testing.T) {
	// run()'s test harness feeds stdin from a bytes.Reader, not a real
	// console, so stty should report the "not a console" failure path.
	code, _, errOut := run(t, "stty")
	if code == 0 {
		t.Error("expected nonzero exit for non-console stdin")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
