package commands

import "testing"

func TestNiceRunsCommand(t *testing.T) {
	code, _, errOut := run(t, "nice", "-n", "10", "cmd", "/c", "exit", "0")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}

func TestNiceMissingCommand(t *testing.T) {
	code, _, errOut := run(t, "nice", "-n", "10")
	if code == 0 {
		t.Error("expected nonzero exit for missing command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestReniceMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "renice", "10")
	if code == 0 {
		t.Error("expected nonzero exit when -p PID is missing")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestReniceInvalidPID(t *testing.T) {
	code, _, errOut := run(t, "renice", "10", "-p", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for an invalid PID")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
