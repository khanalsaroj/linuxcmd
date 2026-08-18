package commands

import (
	"path/filepath"
	"strings"
	"testing"
)

// These commands read live machine state, so the assertions are
// structural: the command must succeed and produce the shape of output
// callers parse, without depending on how many disks or adapters the
// machine running the tests happens to have.

func TestIfconfigListsInterfaces(t *testing.T) {
	code, out, errOut := run(t, "ifconfig", "-a")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "flags=<") {
		t.Errorf("ifconfig output = %q, want net-tools style flags", out)
	}
	if !strings.Contains(out, "mtu ") {
		t.Errorf("ifconfig output = %q, want an mtu field", out)
	}
}

func TestIfconfigShortHeader(t *testing.T) {
	code, out, errOut := run(t, "ifconfig", "-s")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, column := range []string{"Iface", "MTU", "RX-OK", "TX-OK"} {
		if !strings.Contains(out, column) {
			t.Errorf("ifconfig -s output is missing the %s column:\n%s", column, out)
		}
	}
}

func TestIfconfigRejectsUnknownInterface(t *testing.T) {
	code, _, errOut := run(t, "ifconfig", "definitely-not-an-adapter")
	if code == 0 {
		t.Error("expected a nonzero exit for an unknown interface")
	}
	if !strings.Contains(errOut, "does not exist") {
		t.Errorf("stderr = %q, want a does-not-exist message", errOut)
	}
}

func TestLsofListsEndpoints(t *testing.T) {
	code, out, errOut := run(t, "lsof", "-i")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, column := range []string{"COMMAND", "PID", "TYPE", "NODE", "NAME"} {
		if !strings.Contains(out, column) {
			t.Errorf("lsof output is missing the %s column:\n%s", column, out)
		}
	}
}

// Every machine has something listening, but not on a predictable port,
// so this checks that a filter narrows the result rather than that a
// specific port is present.
func TestLsofPortFilterNarrowsOutput(t *testing.T) {
	_, all, _ := run(t, "lsof", "-i")
	_, filtered, _ := run(t, "lsof", "-i:65533")
	if len(filtered) > len(all) {
		t.Errorf("a port filter produced more output than no filter at all")
	}
}

func TestLsofMatchesFilter(t *testing.T) {
	e := lsofEndpoint{Local: "0.0.0.0:8080", Remote: "10.0.0.1:443"}
	tests := []struct {
		filter string
		want   bool
	}{
		{"", true},
		{":8080", true},
		{":443", true},
		{":9999", false},
		{"@10.0.0.1", true},
		{"@192.168.1.1", false},
	}
	for _, tt := range tests {
		if got := lsofMatchesFilter(e, tt.filter); got != tt.want {
			t.Errorf("lsofMatchesFilter(%q) = %v, want %v", tt.filter, got, tt.want)
		}
	}
}

func TestLsofTerseListsOnlyPids(t *testing.T) {
	code, out, errOut := run(t, "lsof", "-t", "-i")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, line := range strings.Fields(out) {
		for _, r := range line {
			if r < '0' || r > '9' {
				t.Fatalf("lsof -t emitted a non-numeric line: %q", line)
			}
		}
	}
}

func TestLsblkListsVolumes(t *testing.T) {
	code, out, errOut := run(t, "lsblk")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, column := range []string{"NAME", "SIZE", "TYPE", "MOUNTPOINT"} {
		if !strings.Contains(out, column) {
			t.Errorf("lsblk output is missing the %s column:\n%s", column, out)
		}
	}
	// The volume the tests are running from must appear somewhere.
	if !strings.Contains(out, ":") {
		t.Errorf("lsblk output = %q, want at least one drive letter", out)
	}
}

func TestLsblkFilesystemView(t *testing.T) {
	code, out, errOut := run(t, "lsblk", "-f")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "FSTYPE") {
		t.Errorf("lsblk -f output = %q, want an FSTYPE column", out)
	}
}

func TestBlkidReportsTypeAndUuid(t *testing.T) {
	code, out, errOut := run(t, "blkid")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, `TYPE="`) {
		t.Errorf("blkid output = %q, want a TYPE tag", out)
	}
	if !strings.Contains(out, `UUID="`) {
		t.Errorf("blkid output = %q, want a UUID tag", out)
	}
}

func TestBlkidExitsTwoWhenNothingMatches(t *testing.T) {
	code, _, _ := run(t, "blkid", "QQ:")
	if code != 2 {
		t.Errorf("exit code = %d, want 2 when no device matches", code)
	}
}

func TestFormatVolumeSerial(t *testing.T) {
	if got := formatVolumeSerial(0x1A2B3C4D); got != "1A2B-3C4D" {
		t.Errorf("formatVolumeSerial = %q, want %q", got, "1A2B-3C4D")
	}
}

func TestVmstatProducesOneSample(t *testing.T) {
	code, out, errOut := run(t, "vmstat")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("vmstat produced %d lines, want 2 header lines and 1 sample:\n%s", len(lines), out)
	}
	if !strings.Contains(lines[1], "swpd") || !strings.Contains(lines[1], "free") {
		t.Errorf("vmstat column header = %q, want vmstat's own columns", lines[1])
	}
	if len(strings.Fields(lines[2])) != 17 {
		t.Errorf("vmstat sample has %d fields, want 17: %q", len(strings.Fields(lines[2])), lines[2])
	}
}

func TestVmstatRejectsBadDelay(t *testing.T) {
	code, _, errOut := run(t, "vmstat", "not-a-number")
	if code == 0 {
		t.Error("expected a nonzero exit for an invalid delay")
	}
	if !strings.Contains(errOut, "invalid delay") {
		t.Errorf("stderr = %q, want an invalid-delay message", errOut)
	}
}

func TestIostatReportsCpu(t *testing.T) {
	code, out, errOut := run(t, "iostat", "-c")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "avg-cpu:") {
		t.Errorf("iostat -c output = %q, want the avg-cpu block", out)
	}
	// iowait and steal have no Windows counterpart and must read zero
	// rather than carry an invented value.
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 6 && strings.Contains(line, ".") {
			if fields[3] != "0.00" || fields[4] != "0.00" {
				t.Errorf("iowait/steal = %s/%s, want 0.00 on Windows", fields[3], fields[4])
			}
		}
	}
}

func TestIostatReportsDisks(t *testing.T) {
	code, out, errOut := run(t, "iostat", "-d")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Device") || !strings.Contains(out, "tps") {
		t.Errorf("iostat -d output = %q, want the device table header", out)
	}
	if strings.Contains(out, "avg-cpu") {
		t.Errorf("iostat -d should omit the CPU block:\n%s", out)
	}
}

func TestGetfaclShowsEntriesForAFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acl.txt")
	mustWriteFile(t, path, "x")

	code, out, errOut := run(t, "getfacl", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "# file: ") {
		t.Errorf("getfacl output = %q, want a '# file:' header", out)
	}
	if !strings.Contains(out, "# owner: ") {
		t.Errorf("getfacl output = %q, want an '# owner:' header", out)
	}
	// The output must not read as a POSIX ACL, because it is not one.
	if !strings.Contains(out, "Windows ACL") {
		t.Errorf("getfacl output = %q, want the note distinguishing it from a POSIX ACL", out)
	}
	if !strings.Contains(out, "user:") && !strings.Contains(out, "group:") {
		t.Errorf("getfacl output = %q, want at least one access entry", out)
	}
}

func TestGetfaclOmitHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "acl.txt")
	mustWriteFile(t, path, "x")

	code, out, errOut := run(t, "getfacl", "-c", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if strings.Contains(out, "# file:") {
		t.Errorf("getfacl -c output = %q, want no header", out)
	}
}

func TestGetfaclReportsMissingFile(t *testing.T) {
	code, _, errOut := run(t, "getfacl", filepath.Join(t.TempDir(), "absent.txt"))
	if code == 0 {
		t.Error("expected a nonzero exit for a missing file")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestMaskToRWX(t *testing.T) {
	tests := []struct {
		mask uint32
		want string
	}{
		{fileReadData, "r--"},
		{fileWriteData, "-w-"},
		{fileExecute, "--x"},
		{fileReadData | fileWriteData | fileExecute, "rwx"},
		{genericAll, "rwx"},
		{genericRead, "r--"},
		{0, "---"},
	}
	for _, tt := range tests {
		if got := maskToRWX(tt.mask); got != tt.want {
			t.Errorf("maskToRWX(%#x) = %q, want %q", tt.mask, got, tt.want)
		}
	}
}

// Rights that rwx cannot express must be named rather than silently
// dropped, so an entry is never shown as granting less than it does.
func TestMaskExtrasNamesUnrepresentableRights(t *testing.T) {
	got := maskExtras(writeDAC | writeOwner)
	if !strings.Contains(got, "change-permissions") {
		t.Errorf("maskExtras = %q, want it to name WRITE_DAC", got)
	}
	if !strings.Contains(got, "take-ownership") {
		t.Errorf("maskExtras = %q, want it to name WRITE_OWNER", got)
	}
	if maskExtras(fileReadData) != "" {
		t.Errorf("maskExtras(FILE_READ_DATA) = %q, want empty", maskExtras(fileReadData))
	}
}
