package commands

import (
	"bytes"
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

func runWithStdin(t *testing.T, name string, stdin string, args ...string) (exitCode int, stdout, stderr string) {
	t.Helper()
	cmd, ok := command.Lookup(name)
	if !ok {
		t.Fatalf("command %q is not registered", name)
	}
	var outBuf, errBuf bytes.Buffer
	ctx := &command.Context{
		CommandName: name,
		Args:        args,
		Stdin:       strings.NewReader(stdin),
		Stdout:      &outBuf,
		Stderr:      &errBuf,
	}
	exitCode = cmd.Run(ctx)
	return exitCode, outBuf.String(), errBuf.String()
}

func TestTRTranslate(t *testing.T) {
	code, out, errOut := runWithStdin(t, "tr", "hello", "a-z", "A-Z")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "HELLO" {
		t.Errorf("tr a-z A-Z output = %q, want %q", out, "HELLO")
	}
}

func TestTRDelete(t *testing.T) {
	code, out, errOut := runWithStdin(t, "tr", "hello world", "-d", "lo")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "he wrd" {
		t.Errorf("tr -d lo output = %q, want %q", out, "he wrd")
	}
}

func TestTRSqueeze(t *testing.T) {
	code, out, errOut := runWithStdin(t, "tr", "aaabbbccc", "-s", "a-c")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "abc" {
		t.Errorf("tr -s a-c output = %q, want %q", out, "abc")
	}
}

func TestTRMissingArgs(t *testing.T) {
	code, _, errOut := runWithStdin(t, "tr", "x")
	if code == 0 {
		t.Error("expected nonzero exit for missing arguments")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
