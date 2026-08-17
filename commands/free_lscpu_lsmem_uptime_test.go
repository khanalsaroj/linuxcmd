package commands

import (
	"strings"
	"testing"
)

func TestFreeReportsMemory(t *testing.T) {
	code, out, errOut := run(t, "free")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Mem:") {
		t.Errorf("free output = %q, want it to contain 'Mem:'", out)
	}
}

func TestFreeHumanReadable(t *testing.T) {
	code, out, errOut := run(t, "free", "-h")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected non-empty output")
	}
}

func TestLscpuReportsArchitecture(t *testing.T) {
	code, out, errOut := run(t, "lscpu")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Architecture:") || !strings.Contains(out, "CPU(s):") {
		t.Errorf("lscpu output = %q, want Architecture and CPU(s) lines", out)
	}
}

func TestLsmemReportsTotals(t *testing.T) {
	code, out, errOut := run(t, "lsmem")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Total online memory:") {
		t.Errorf("lsmem output = %q, want a total memory line", out)
	}
}

func TestUptimeReportsUpTime(t *testing.T) {
	code, out, errOut := run(t, "uptime")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.HasPrefix(out, "up ") {
		t.Errorf("uptime output = %q, want it to start with 'up '", out)
	}
}
