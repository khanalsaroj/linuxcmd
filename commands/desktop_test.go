package commands

import (
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

// These commands all have real side effects -- launching an application,
// raising a UAC prompt, attaching a network drive -- so the tests
// exercise argument handling and the refusal paths, and never the paths
// that would actually act on the machine.

func TestXdgOpenRequiresAnOperand(t *testing.T) {
	for _, name := range []string{"xdg-open", "open"} {
		code, _, errOut := run(t, name)
		if code != xdgExitSyntax {
			t.Errorf("%s: exit code = %d, want %d for a syntax error", name, code, xdgExitSyntax)
		}
		if !strings.Contains(errOut, "usage:") {
			t.Errorf("%s: stderr = %q, want a usage line", name, errOut)
		}
	}
}

// xdg-open's documented exit codes are what scripts branch on, so a
// missing file must produce 2 rather than a generic failure.
func TestXdgOpenMissingFileUsesSpecifiedExitCode(t *testing.T) {
	code, _, errOut := run(t, "xdg-open", "no-such-file-anywhere.qqq")
	if code != xdgExitNotFound {
		t.Errorf("exit code = %d, want %d for a missing file", code, xdgExitNotFound)
	}
	if !strings.Contains(errOut, "No such file or directory") {
		t.Errorf("stderr = %q, want a Linux-style not-found message", errOut)
	}
}

// A Windows drive path must not be mistaken for a URI scheme, or every
// absolute path would bypass the existence check.
func TestHasURIScheme(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"mailto:someone@example.com", true},
		{"ms-settings:display", true},
		{`C:\Users\x`, false},
		{"C:/Users/x", false},
		{"d:/data", false},
		{"./relative/path", false},
		{"plain.txt", false},
		{"no-scheme-here", false},
	}
	for _, tt := range tests {
		if got := hasURIScheme(tt.input); got != tt.want {
			t.Errorf("hasURIScheme(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestClipboardRoundTrip(t *testing.T) {
	const want = "linuxcmd clipboard round trip"

	code, _, errOut := runWithStdin(t, "xclip", want)
	if code != 0 {
		// A session without an interactive window station (some CI
		// configurations) has no clipboard at all; that is an
		// environment limitation, not a defect.
		t.Skipf("clipboard is unavailable in this environment: %s", errOut)
	}

	code, out, errOut := run(t, "xclip", "-o")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != want {
		t.Errorf("clipboard round trip = %q, want %q", out, want)
	}

	// xsel with no direction flag reads, where xclip writes.
	code, out, errOut = run(t, "xsel")
	if code != 0 {
		t.Fatalf("xsel: exit code = %d, stderr = %q", code, errOut)
	}
	if out != want {
		t.Errorf("xsel read = %q, want %q", out, want)
	}
}

func TestClipboardTrimsTrailingNewline(t *testing.T) {
	code, _, errOut := runWithStdin(t, "xclip", "trimmed\n", "-r")
	if code != 0 {
		t.Skipf("clipboard is unavailable in this environment: %s", errOut)
	}
	_, out, _ := run(t, "xclip", "-o")
	if out != "trimmed" {
		t.Errorf("xclip -r stored %q, want the trailing newline removed", out)
	}
}

func TestMountListsVolumes(t *testing.T) {
	code, out, errOut := run(t, "mount")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, " on ") || !strings.Contains(out, " type ") {
		t.Errorf("mount output = %q, want Linux mount's 'X on Y type Z' layout", out)
	}
}

// Mounting anything other than a network share has no Windows
// equivalent and must be refused with a pointer to the right tool.
func TestMountRefusesNonShareSources(t *testing.T) {
	code, _, errOut := run(t, "mount", "/dev/sda1", "/mnt/data")
	if code == 0 {
		t.Error("expected a nonzero exit for a non-share source")
	}
	if !strings.Contains(errOut, "only network shares can be mounted") {
		t.Errorf("stderr = %q, want a clear refusal", errOut)
	}
	if !strings.Contains(errOut, "Mount-DiskImage") {
		t.Errorf("stderr = %q, want a pointer to the tool that can do this", errOut)
	}
}

func TestMountRejectsWrongOperandCount(t *testing.T) {
	code, _, errOut := run(t, "mount", "//server/share")
	if code != command.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, command.ExitUsage)
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want a usage line", errOut)
	}
}

func TestUmountRequiresADriveLetter(t *testing.T) {
	code, _, errOut := run(t, "umount", "/mnt/data")
	if code == 0 {
		t.Error("expected a nonzero exit for something that is not a drive letter")
	}
	if !strings.Contains(errOut, "not a drive letter") {
		t.Errorf("stderr = %q, want a clear message", errOut)
	}
}

func TestUmountRequiresAnOperand(t *testing.T) {
	code, _, errOut := run(t, "umount")
	if code != command.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, command.ExitUsage)
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want a usage line", errOut)
	}
}

func TestSudoRequiresACommand(t *testing.T) {
	code, _, errOut := run(t, "sudo")
	if code != command.ExitUsage {
		t.Errorf("exit code = %d, want %d", code, command.ExitUsage)
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want a usage line", errOut)
	}
}

// sudo -v asks about a credential cache, which Windows does not have.
// Saying so is more useful than pretending the cache was refreshed.
func TestSudoValidateExplainsNoCredentialCache(t *testing.T) {
	code, _, errOut := run(t, "sudo", "-v")
	if code != command.ExitSuccess {
		t.Errorf("exit code = %d, want success", code)
	}
	if !strings.Contains(errOut, "no credential cache") {
		t.Errorf("stderr = %q, want an explanation of the difference", errOut)
	}
}

// "sudo rm ..." must elevate linuxcmd's own rm, since there is no rm.exe
// on a stock Windows system.
func TestResolveElevationTargetFindsRegisteredCommand(t *testing.T) {
	target, args, err := resolveElevationTarget([]string{"rev", "-x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == "" {
		t.Error("expected a target executable")
	}
	if len(args) != 2 || args[0] != "rev" || args[1] != "-x" {
		t.Errorf("args = %v, want the command name followed by its arguments", args)
	}
}

func TestResolveElevationTargetRejectsUnknownCommand(t *testing.T) {
	_, _, err := resolveElevationTarget([]string{"definitely-not-a-command-xyz"})
	if err == nil {
		t.Error("expected an error for an unknown command")
	}
	if !strings.Contains(err.Error(), "command not found") {
		t.Errorf("error = %v, want a command-not-found message", err)
	}
}

// ShellExecute takes parameters as one string, so the splitting Windows
// did on the way in has to be undone correctly on the way out.
func TestQuoteCommandLine(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"a", "b"}, "a b"},
		{[]string{"with space"}, `"with space"`},
		{[]string{"plain", "two words"}, `plain "two words"`},
		{[]string{""}, `""`},
		{[]string{`quote"inside`}, `"quote\"inside"`},
	}
	for _, tt := range tests {
		if got := quoteCommandLine(tt.args); got != tt.want {
			t.Errorf("quoteCommandLine(%v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}
