package commands

import "testing"

func TestTrueSucceeds(t *testing.T) {
	code, _, _ := run(t, "true")
	if code != 0 {
		t.Errorf("expected true to exit 0, got %d", code)
	}
}

func TestFalseFails(t *testing.T) {
	code, _, _ := run(t, "false")
	if code == 0 {
		t.Errorf("expected false to exit nonzero")
	}
}
