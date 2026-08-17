package commands

import "testing"

// pkill's matching kills real OS processes by name, so tests intentionally
// avoid exercising the actual-termination path (it could kill unrelated
// processes on the machine running the tests) and only cover the safe,
// read-only "no match" case.
func TestPkillNoMatch(t *testing.T) {
	code, _, _ := run(t, "pkill", "definitely-not-a-real-process-xyz")
	if code == 0 {
		t.Error("expected nonzero exit for no matches")
	}
}

func TestPkillMissingArgument(t *testing.T) {
	code, _, errOut := run(t, "pkill")
	if code == 0 {
		t.Error("expected nonzero exit for missing pattern")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
