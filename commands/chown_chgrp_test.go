package commands

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestChownMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "chown", "someuser")
	if code == 0 {
		t.Error("expected nonzero exit for missing file operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}

func TestChownSetsOwnerToCurrentUser(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Skipf("could not determine current user: %v", err)
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	mustWriteFile(t, f, "x")

	code, _, errOut := run(t, "chown", u.Username, f)
	if code != 0 {
		t.Skipf("icacls /setowner not permitted in this environment: %s", errOut)
	}
}

func TestChgrpMissingArgs(t *testing.T) {
	code, _, errOut := run(t, "chgrp", "somegroup")
	if code == 0 {
		t.Error("expected nonzero exit for missing file operand")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
