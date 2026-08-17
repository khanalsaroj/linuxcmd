package commands

import (
	"os"
	"strings"
	"testing"
)

func TestMktempCreatesFile(t *testing.T) {
	code, out, errOut := run(t, "mktemp")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	name := strings.TrimSpace(out)
	defer os.Remove(name)
	if _, err := os.Stat(name); err != nil {
		t.Errorf("expected mktemp to create %q: %v", name, err)
	}
}

func TestMktempDirectory(t *testing.T) {
	code, out, errOut := run(t, "mktemp", "-d")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	name := strings.TrimSpace(out)
	defer os.RemoveAll(name)
	info, err := os.Stat(name)
	if err != nil {
		t.Fatalf("expected mktemp -d to create %q: %v", name, err)
	}
	if !info.IsDir() {
		t.Error("expected mktemp -d to create a directory")
	}
}
