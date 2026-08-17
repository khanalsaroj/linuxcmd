package commands

import (
	"strings"
	"testing"
)

func TestHelpListsCommands(t *testing.T) {
	code, out, errOut := run(t, "help")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "cat") || !strings.Contains(out, "grep") {
		t.Errorf("help output = %q, want it to list registered commands", out)
	}
}

func TestHelpSpecificCommand(t *testing.T) {
	code, out, errOut := run(t, "help", "cat")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "cat") {
		t.Errorf("help cat output = %q, want it to describe 'cat'", out)
	}
}

func TestHelpUnknownCommand(t *testing.T) {
	code, _, errOut := run(t, "help", "not-a-real-command")
	if code == 0 {
		t.Error("expected nonzero exit for an unknown command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
