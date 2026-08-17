package commands

import (
	"net"
	"strings"
	"testing"
)

func requireDNS(t *testing.T) {
	t.Helper()
	if _, err := net.LookupHost("example.com"); err != nil {
		t.Skipf("no DNS connectivity in this environment: %v", err)
	}
}

func TestHostResolvesExampleCom(t *testing.T) {
	requireDNS(t)
	code, out, errOut := run(t, "host", "example.com")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "example.com") {
		t.Errorf("host output = %q, want it to mention example.com", out)
	}
}

func TestNslookupResolvesExampleCom(t *testing.T) {
	requireDNS(t)
	code, out, errOut := run(t, "nslookup", "example.com")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "Address:") {
		t.Errorf("nslookup output = %q, want an Address: line", out)
	}
}

func TestDigResolvesExampleCom(t *testing.T) {
	requireDNS(t)
	code, out, errOut := run(t, "dig", "example.com")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "ANSWER SECTION") {
		t.Errorf("dig output = %q, want an ANSWER SECTION", out)
	}
}

func TestDigUnsupportedType(t *testing.T) {
	code, _, errOut := run(t, "dig", "example.com", "BOGUS")
	if code == 0 {
		t.Error("expected nonzero exit for an unsupported record type")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
