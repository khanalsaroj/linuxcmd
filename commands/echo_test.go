package commands

import (
	"os"
	"strings"
	"testing"
)

func TestEchoBasic(t *testing.T) {
	code, out, errOut := run(t, "echo", "hello", "world")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello world\n" {
		t.Errorf("echo output = %q, want %q", out, "hello world\n")
	}
}

func TestEchoNoNewlineFlag(t *testing.T) {
	code, out, errOut := run(t, "echo", "-n", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello" {
		t.Errorf("echo -n output = %q, want %q", out, "hello")
	}
}

func TestEchoExpandsHome(t *testing.T) {
	home := os.Getenv("USERPROFILE")
	code, out, errOut := run(t, "echo", "$HOME")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != home {
		t.Errorf("echo $HOME = %q, want %q", strings.TrimSpace(out), home)
	}
}

func TestEchoEscapes(t *testing.T) {
	code, out, errOut := run(t, "echo", "-e", `a\tb\nc`)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "a\tb\nc\n"
	if out != want {
		t.Errorf("echo -e output = %q, want %q", out, want)
	}
}

func TestEchoEmpty(t *testing.T) {
	code, out, errOut := run(t, "echo")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "\n" {
		t.Errorf("bare echo output = %q, want a single newline", out)
	}
}

func TestWhoami(t *testing.T) {
	code, out, errOut := run(t, "whoami")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected a non-empty username")
	}
	if strings.Contains(out, `\`) {
		t.Errorf("expected DOMAIN\\ prefix stripped, got %q", out)
	}
}

func TestHostname(t *testing.T) {
	code, out, errOut := run(t, "hostname")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected a non-empty hostname")
	}
}
