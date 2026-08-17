package commands

import "testing"

// Reading the Security event log typically requires administrator
// rights, so this only checks that the command completes (wraps
// wevtutil.exe correctly) rather than asserting success.
func TestLastRuns(t *testing.T) {
	run(t, "last")
}
