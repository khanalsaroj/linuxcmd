package commands

import (
	"strings"
	"testing"
)

func TestXargsBuildsCommandFromStdin(t *testing.T) {
	code, out, errOut := runWithStdin(t, "xargs", "one two three", "cmd", "/c", "echo")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") || !strings.Contains(out, "three") {
		t.Errorf("xargs output = %q, want it to contain all three tokens", out)
	}
}

func TestXargsBatchSize(t *testing.T) {
	code, out, errOut := runWithStdin(t, "xargs", "a b c", "-n", "1", "cmd", "/c", "echo")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Count(out, "\n") < 3 {
		t.Errorf("xargs -n 1 output = %q, want one invocation per token", out)
	}
}

func TestXargsEmptyStdin(t *testing.T) {
	code, out, errOut := runWithStdin(t, "xargs", "", "cmd", "/c", "echo", "hi")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "" {
		t.Errorf("expected no invocation for empty stdin, got %q", out)
	}
}
