package commands

import (
	"strings"
	"testing"
)

func TestBasenameStripsDirectory(t *testing.T) {
	code, out, errOut := run(t, "basename", `C:\work\a.txt`)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "a.txt" {
		t.Errorf("basename output = %q, want %q", strings.TrimSpace(out), "a.txt")
	}
}

func TestBasenameStripsSuffix(t *testing.T) {
	code, out, errOut := run(t, "basename", `C:\work\a.txt`, ".txt")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "a" {
		t.Errorf("basename output = %q, want %q", strings.TrimSpace(out), "a")
	}
}

func TestDirnamePrintsDirectory(t *testing.T) {
	code, out, errOut := run(t, "dirname", `C:\work\a.txt`)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != `C:\work` {
		t.Errorf("dirname output = %q, want %q", strings.TrimSpace(out), `C:\work`)
	}
}
