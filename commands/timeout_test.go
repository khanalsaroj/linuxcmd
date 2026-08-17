package commands

import "testing"

func TestTimeoutRunsCommandToCompletion(t *testing.T) {
	code, _, errOut := run(t, "timeout", "5", "cmd", "/c", "exit", "0")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}

func TestTimeoutKillsSlowCommand(t *testing.T) {
	code, _, errOut := run(t, "timeout", "0.2", "cmd", "/c", "ping", "-n", "20", "127.0.0.1", ">nul")
	if code != 124 {
		t.Fatalf("exit code = %d, want 124, stderr = %q", code, errOut)
	}
}

func TestTimeoutMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "timeout")
	if code == 0 {
		t.Error("expected nonzero exit for missing arguments")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
