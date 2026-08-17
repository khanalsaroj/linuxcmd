package commands

import "testing"

// Reading these logs may or may not require elevated rights depending on
// the account running the tests, so these only check that the commands
// wrap wevtutil.exe and complete rather than asserting success.
func TestDmesgRuns(t *testing.T) {
	run(t, "dmesg")
}

func TestJournalctlRuns(t *testing.T) {
	run(t, "journalctl", "-n", "5")
}

func TestJournalctlInvalidCount(t *testing.T) {
	code, _, errOut := run(t, "journalctl", "-n", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for an invalid -n value")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
