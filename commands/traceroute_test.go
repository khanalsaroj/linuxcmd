package commands

import "testing"

// A real traceroute run against an actual host can take tens of seconds
// (multiple hops, each with its own timeout), so this only covers the
// fast, deterministic usage-error path.
func TestTracerouteMissingHost(t *testing.T) {
	code, _, errOut := run(t, "traceroute")
	if code == 0 {
		t.Error("expected nonzero exit for missing host")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
