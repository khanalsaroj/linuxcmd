package commands

import (
	"strings"
	"testing"
	"time"
)

func TestDateDefaultOutputNotEmpty(t *testing.T) {
	code, out, errOut := run(t, "date")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty date output")
	}
}

func TestDateCustomFormat(t *testing.T) {
	code, out, errOut := run(t, "date", "+%Y-%m-%d")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := time.Now().Format("2006-01-02")
	if strings.TrimSpace(out) != want {
		t.Errorf("date +%%Y-%%m-%%d output = %q, want %q", strings.TrimSpace(out), want)
	}
}

func TestDateUTCFlag(t *testing.T) {
	code, out, errOut := run(t, "date", "-u", "+%Z")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) != "UTC" {
		t.Errorf("date -u +%%Z output = %q, want %q", strings.TrimSpace(out), "UTC")
	}
}
