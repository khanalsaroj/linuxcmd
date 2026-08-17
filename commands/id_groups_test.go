package commands

import (
	"strings"
	"testing"
)

func TestIdPrintsUidGid(t *testing.T) {
	code, out, errOut := run(t, "id")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.HasPrefix(out, "uid=") || !strings.Contains(out, "gid=") {
		t.Errorf("id output = %q, want it to start with uid= and contain gid=", out)
	}
}

func TestGroupsPrintsSomething(t *testing.T) {
	code, out, errOut := run(t, "groups")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty groups output")
	}
}
