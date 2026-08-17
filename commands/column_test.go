package commands

import (
	"strings"
	"testing"
)

func TestColumnAlignsWhitespaceFields(t *testing.T) {
	code, out, errOut := runWithStdin(t, "column", "a bb ccc\nlonger short x\n", "-t")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %q", out)
	}
	col2Start0 := strings.Index(lines[0], "bb")
	col2Start1 := strings.Index(lines[1], "short")
	if col2Start0 != col2Start1 {
		t.Errorf("expected second column to start at the same offset, got %q / %q", lines[0], lines[1])
	}
}

func TestColumnCustomSeparator(t *testing.T) {
	code, out, errOut := runWithStdin(t, "column", "a,b,c\n", "-s,")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") || !strings.Contains(out, "c") {
		t.Errorf("column -s, output = %q", out)
	}
}
