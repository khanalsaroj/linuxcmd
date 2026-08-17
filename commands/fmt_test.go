package commands

import (
	"strings"
	"testing"
)

func TestFmtWrapsToWidth(t *testing.T) {
	text := "one two three four five six seven eight nine ten"
	code, out, errOut := runWithStdin(t, "fmt", text, "-w", "15")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if len(line) > 15 {
			t.Errorf("line exceeds width 15: %q", line)
		}
	}
	if !strings.Contains(out, "one") || !strings.Contains(out, "ten") {
		t.Errorf("expected all words preserved, got %q", out)
	}
}

func TestFmtPreservesParagraphBreaks(t *testing.T) {
	text := "para one\n\npara two\n"
	code, out, errOut := runWithStdin(t, "fmt", text)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "\n\n") {
		t.Errorf("expected blank line between paragraphs, got %q", out)
	}
}
