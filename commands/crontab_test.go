package commands

import "testing"

func TestCrontabList(t *testing.T) {
	code, _, errOut := run(t, "crontab", "-l")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
}

func TestCrontabMissingFlag(t *testing.T) {
	code, _, errOut := run(t, "crontab")
	if code == 0 {
		t.Error("expected nonzero exit for missing -l")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
