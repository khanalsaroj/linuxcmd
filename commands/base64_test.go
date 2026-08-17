package commands

import (
	"strings"
	"testing"
)

func TestBase64Encode(t *testing.T) {
	code, out, errOut := runWithStdin(t, "base64", "hello")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "aGVsbG8=" {
		t.Errorf("base64 output = %q, want %q", strings.TrimSpace(out), "aGVsbG8=")
	}
}

func TestBase64Decode(t *testing.T) {
	code, out, errOut := runWithStdin(t, "base64", "aGVsbG8=\n", "-d")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello" {
		t.Errorf("base64 -d output = %q, want %q", out, "hello")
	}
}

func TestBase64DecodeInvalid(t *testing.T) {
	code, _, errOut := runWithStdin(t, "base64", "not valid base64!!", "-d")
	if code == 0 {
		t.Error("expected nonzero exit for invalid base64 input")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
