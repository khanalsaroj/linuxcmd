package commands

import (
	"strings"
	"testing"
)

func TestNohupRunsCommand(t *testing.T) {
	// ctx.Stdout in the test harness isn't a real console, so nohup takes
	// its non-interactive path and writes straight through rather than
	// creating nohup.out.
	code, out, errOut := run(t, "nohup", "cmd", "/c", "echo", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("nohup output = %q, want it to contain 'hello'", out)
	}
}

func TestNohupMissingCommand(t *testing.T) {
	code, _, errOut := run(t, "nohup")
	if code == 0 {
		t.Error("expected nonzero exit for missing command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
