package output

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxErrorTextNotExist(t *testing.T) {
	_, err := os.Open(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	if err == nil {
		t.Fatal("expected an error opening a nonexistent file")
	}
	if got := LinuxErrorText(err); got != "No such file or directory" {
		t.Errorf("LinuxErrorText = %q, want %q", got, "No such file or directory")
	}
}

func TestLinuxErrorTextWrappedFsError(t *testing.T) {
	err := &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist}
	if got := LinuxErrorText(err); got != "No such file or directory" {
		t.Errorf("LinuxErrorText = %q, want %q", got, "No such file or directory")
	}
}

func TestLinuxErrorTextUnknownFallsBackToMessage(t *testing.T) {
	custom := errors.New("some very specific windows detail")
	if got := LinuxErrorText(custom); got != custom.Error() {
		t.Errorf("LinuxErrorText should fall back to the original message, got %q", got)
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{500, "500"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{10 * 1024, "10K"},
		{1024 * 1024, "1.0M"},
	}
	for _, c := range cases {
		if got := HumanSize(c.in); got != c.want {
			t.Errorf("HumanSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatModeDirectory(t *testing.T) {
	dir := t.TempDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	mode := FormatMode(info)
	if mode[0] != 'd' {
		t.Errorf("FormatMode for a directory should start with 'd', got %q", mode)
	}
	if len(mode) != 10 {
		t.Errorf("FormatMode should be 10 characters, got %q (%d)", mode, len(mode))
	}
}

func TestFormatModeFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	mode := FormatMode(info)
	if mode[0] != '-' {
		t.Errorf("FormatMode for a regular file should start with '-', got %q", mode)
	}
}
