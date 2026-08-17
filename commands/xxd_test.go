package commands

import (
	"strings"
	"testing"
)

func TestXxdDump(t *testing.T) {
	code, out, errOut := runWithStdin(t, "xxd", "AB")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "00000000") || !strings.Contains(out, "4142") {
		t.Errorf("xxd output = %q, want offset and hex bytes for 'AB'", out)
	}
	if !strings.Contains(out, "AB") {
		t.Errorf("xxd output = %q, want ASCII gutter showing 'AB'", out)
	}
}

func TestXxdRoundTrip(t *testing.T) {
	_, dump, errOut := runWithStdin(t, "xxd", "hello world")
	if errOut != "" {
		t.Fatalf("unexpected stderr producing dump: %q", errOut)
	}
	code, out, errOut := runWithStdin(t, "xxd", dump, "-r")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello world" {
		t.Errorf("xxd -r round trip = %q, want %q", out, "hello world")
	}
}
