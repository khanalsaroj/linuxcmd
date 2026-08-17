package commands

import "testing"

// nc's Run blocks until data stops flowing in either direction, which
// makes full bidirectional-copy behavior awkward to assert deterministically
// in a unit test. This covers the fast, deterministic error paths instead.
func TestNcConnectionRefused(t *testing.T) {
	code, _, errOut := run(t, "nc", "-w", "1", "127.0.0.1", "1")
	if code == 0 {
		t.Error("expected nonzero exit connecting to a closed port")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestNcMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "nc")
	if code == 0 {
		t.Error("expected nonzero exit for missing arguments")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
