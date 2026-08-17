package commands

import (
	"strings"
	"testing"
)

func TestStringsExtractsPrintableSequences(t *testing.T) {
	data := "\x00\x01hello\x00\x02world\x03ab\x00"
	code, out, errOut := runWithStdin(t, "strings", data)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("strings output = %q, want hello and world", out)
	}
}

func TestStringsMinLength(t *testing.T) {
	code, out, errOut := runWithStdin(t, "strings", "\x00ab\x00abcdef\x00", "-n", "5")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "ab\n") {
		t.Errorf("expected short sequence to be dropped, got %q", out)
	}
	if !strings.Contains(out, "abcdef") {
		t.Errorf("expected long sequence to be kept, got %q", out)
	}
}
