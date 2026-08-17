package commands

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestOpensslRandHex(t *testing.T) {
	code, out, errOut := run(t, "openssl", "rand", "-hex", "16")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got := strings.TrimSpace(out)
	if len(got) != 32 {
		t.Errorf("openssl rand -hex 16 output = %q, want a 32-character hex string", got)
	}
}

func TestOpensslDgstSha256(t *testing.T) {
	code, out, errOut := runWithStdin(t, "openssl", "hello", "dgst", "-sha256")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := fmt.Sprintf("%x", sha256.Sum256([]byte("hello")))
	if strings.TrimSpace(out) != want {
		t.Errorf("openssl dgst -sha256 output = %q, want %q", strings.TrimSpace(out), want)
	}
}

func TestOpensslUnknownSubcommand(t *testing.T) {
	code, _, errOut := run(t, "openssl", "bogus")
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported subcommand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
