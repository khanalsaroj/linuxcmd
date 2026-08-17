package commands

import (
	"testing"
	"time"
)

func TestSleepWaitsApproximateDuration(t *testing.T) {
	start := time.Now()
	code, _, errOut := run(t, "sleep", "0.05")
	elapsed := time.Since(start)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("sleep returned too early: %v", elapsed)
	}
}

func TestSleepInvalidInterval(t *testing.T) {
	code, _, errOut := run(t, "sleep", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for invalid interval")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestSleepMissingOperand(t *testing.T) {
	code, _, errOut := run(t, "sleep")
	if code == 0 {
		t.Error("expected nonzero exit for missing operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
