package commands

import (
	"path/filepath"
	"testing"
)

func TestJoinOnCommonField(t *testing.T) {
	dir := t.TempDir()
	users := filepath.Join(dir, "users.txt")
	roles := filepath.Join(dir, "roles.txt")
	mustWriteFile(t, users, "1 alice\n2 bob\n")
	mustWriteFile(t, roles, "1 admin\n2 editor\n")

	code, out, errOut := run(t, "join", users, roles)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "1 alice admin\n2 bob editor\n"
	if out != want {
		t.Errorf("join output = %q, want %q", out, want)
	}
}

func TestJoinSkipsUnmatchedKeys(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	mustWriteFile(t, a, "1 alice\n2 bob\n")
	mustWriteFile(t, b, "1 admin\n")

	code, out, errOut := run(t, "join", a, b)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	want := "1 alice admin\n"
	if out != want {
		t.Errorf("join output = %q, want %q", out, want)
	}
}
