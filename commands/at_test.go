package commands

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

var atTaskNameRe = regexp.MustCompile(`task '([^']+)'`)

func TestAtSchedulesAndCleansUp(t *testing.T) {
	// 23:59 is far enough out that it won't fire during the test; the
	// task is deleted again in cleanup regardless.
	code, out, errOut := run(t, "at", "23:59", "cmd", "/c", "exit", "0")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	m := atTaskNameRe.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("could not find scheduled task name in output %q", out)
	}
	taskName := m[1]
	t.Cleanup(func() {
		exec.Command(schtasksExe(), "/delete", "/tn", taskName, "/f").Run()
	})
	if !strings.Contains(out, "job scheduled") {
		t.Errorf("at output = %q, want a scheduling confirmation", out)
	}
}

func TestAtMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "at", "23:59")
	if code == 0 {
		t.Error("expected nonzero exit for missing command")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
