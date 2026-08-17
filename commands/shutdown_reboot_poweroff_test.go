package commands

import "testing"

// shutdown/reboot/poweroff wrap shutdown.exe and would actually restart
// or power off the machine running the tests if the exec.Command path
// were exercised. These tests deliberately stay on the argument-
// validation path (which returns before any process is started) and
// never invoke the real command.

func TestShutdownRequiresTimeArgument(t *testing.T) {
	code, _, errOut := run(t, "shutdown", "-r")
	if code == 0 {
		t.Error("expected nonzero exit for a missing TIME argument")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestShutdownRejectsAbsoluteClockTime(t *testing.T) {
	code, _, errOut := run(t, "shutdown", "-r", "18:00")
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported absolute time")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestShutdownDelaySecondsHelper(t *testing.T) {
	cases := map[string]int{"now": 0, "+2": 120, "45": 45}
	for input, want := range cases {
		got, err := shutdownDelaySeconds(input)
		if err != nil {
			t.Fatalf("shutdownDelaySeconds(%q) error: %v", input, err)
		}
		if got != want {
			t.Errorf("shutdownDelaySeconds(%q) = %d, want %d", input, got, want)
		}
	}
	if _, err := shutdownDelaySeconds("bogus"); err == nil {
		t.Error("expected an error for an invalid time string")
	}
}
