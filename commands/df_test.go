package commands

import (
	"strings"
	"testing"
)

func TestDfListsVolumes(t *testing.T) {
	code, out, errOut := run(t, "df")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Errorf("expected a header plus at least one volume line, got %q", out)
	}
	if !strings.Contains(lines[0], "Filesystem") {
		t.Errorf("expected a header line, got %q", lines[0])
	}
}

func TestDfHumanReadable(t *testing.T) {
	code, out, errOut := run(t, "df", "-h")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty output")
	}
}
