package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

// Every entry in the passthrough table must be reachable by name, or the
// table and the registry have drifted apart.
func TestExternalToolsAreRegistered(t *testing.T) {
	for _, tool := range externalTools {
		c, ok := command.Lookup(tool.name)
		if !ok {
			t.Errorf("external tool %q is not registered", tool.name)
			continue
		}
		if c.Summary() == "" {
			t.Errorf("external tool %q has an empty summary", tool.name)
		}
		if len(tool.candidates) == 0 {
			t.Errorf("external tool %q lists no candidate executables", tool.name)
		}
		if tool.hint == "" {
			t.Errorf("external tool %q has no install hint", tool.name)
		}
		for _, c := range tool.candidates {
			if filepath.Ext(c.File) == "" {
				t.Errorf("external tool %q candidate %q has no file extension", tool.name, c.File)
			}
		}
	}
}

// A tool that is not installed must produce an actionable message rather
// than Windows' "is not recognized as an internal or external command".
func TestExternalToolReportsInstallHint(t *testing.T) {
	tool := externalTool{
		name:       "definitely-not-installed",
		summary:    "test double",
		candidates: exeCandidates("linuxcmd-no-such-tool.exe"),
		hint:       "install the thing",
	}
	var stdout, stderr strings.Builder
	ctx := &command.Context{
		CommandName: tool.name,
		Stdin:       strings.NewReader(""),
		Stdout:      &stdout,
		Stderr:      &stderr,
	}

	code := tool.Run(ctx)
	if code != command.ExitNotFound {
		t.Errorf("exit code = %d, want %d", code, command.ExitNotFound)
	}
	if !strings.Contains(stderr.String(), "command not found") {
		t.Errorf("stderr = %q, want a not-found message", stderr.String())
	}
	if !strings.Contains(stderr.String(), "install the thing") {
		t.Errorf("stderr = %q, want the install hint", stderr.String())
	}
}

// The installer hardlinks this binary to one .exe per command, so a PATH
// lookup for a wrapper's own name can resolve straight back to this
// process. findExternal must reject any candidate that is the same file
// as the running executable, or every wrapper would fork itself forever.
func TestFindExternalSkipsItsOwnBinary(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Skipf("cannot determine the test binary's path: %v", err)
	}
	dir := filepath.Dir(self)
	name := filepath.Base(self)

	found, _, ok := findExternal(exeCandidates(name), func() []string { return []string{dir} })
	if ok {
		t.Errorf("findExternal returned %q, which is linuxcmd itself", found)
	}
}

// The same guard must hold for a hardlink under a different name in a
// different directory, which is exactly what the installer creates. A
// directory comparison alone would miss this; identity comparison does not.
func TestFindExternalSkipsHardlinkOfItself(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Skipf("cannot determine the test binary's path: %v", err)
	}
	// Not t.TempDir(): the hardlink points at the running test binary, so
	// Windows keeps it locked and the automatic cleanup would fail the
	// test after its assertion has already passed.
	dir, err := os.MkdirTemp("", "linuxcmd-hardlink-")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "pretend-tool.exe")
	t.Cleanup(func() {
		os.Remove(link) // locked while the test binary runs; best effort
		os.Remove(dir)
	})
	if err := os.Link(self, link); err != nil {
		t.Skipf("cannot create a hardlink in this environment: %v", err)
	}

	found, _, ok := findExternal(exeCandidates("pretend-tool.exe"), func() []string { return []string{dir} })
	if ok {
		t.Errorf("findExternal returned %q, a hardlink of linuxcmd itself", found)
	}
}

func TestFindExternalPrefersEarlierCandidate(t *testing.T) {
	dir := t.TempDir()
	second := filepath.Join(dir, "second-choice.exe")
	mustWriteFile(t, second, "not a real program")

	found, _, ok := findExternal(
		exeCandidates("first-choice.exe", "second-choice.exe"),
		func() []string { return []string{dir} },
	)
	if !ok {
		t.Fatal("expected the second candidate to be found")
	}
	if found != second {
		t.Errorf("findExternal = %q, want %q", found, second)
	}
}

// The Python launcher needs "-3" prepended to mean python3, so a
// candidate's prefix arguments must come back with it.
func TestFindExternalReturnsPrefixArguments(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "launcher.exe")
	mustWriteFile(t, exe, "not a real program")

	_, prefix, ok := findExternal(
		[]externalCandidate{{File: "launcher.exe", Args: []string{"-3"}}},
		func() []string { return []string{dir} },
	)
	if !ok {
		t.Fatal("expected the candidate to be found")
	}
	if len(prefix) != 1 || prefix[0] != "-3" {
		t.Errorf("prefix args = %v, want [-3]", prefix)
	}
}

func TestSplitOptionalArgFlag(t *testing.T) {
	tests := []struct {
		args      []string
		wantArgs  []string
		wantValue string
	}{
		{[]string{"-i:8080"}, []string{"-i"}, ":8080"},
		{[]string{"-i@host"}, []string{"-i"}, "@host"},
		{[]string{"-i"}, []string{"-i"}, ""},
		{[]string{"-n", "-i:53", "-P"}, []string{"-n", "-i", "-P"}, ":53"},
		{[]string{"-t"}, []string{"-t"}, ""},
	}
	for _, tt := range tests {
		gotArgs, gotValue := splitOptionalArgFlag(tt.args, 'i')
		if strings.Join(gotArgs, " ") != strings.Join(tt.wantArgs, " ") {
			t.Errorf("splitOptionalArgFlag(%v) args = %v, want %v", tt.args, gotArgs, tt.wantArgs)
		}
		if gotValue != tt.wantValue {
			t.Errorf("splitOptionalArgFlag(%v) value = %q, want %q", tt.args, gotValue, tt.wantValue)
		}
	}
}

// nmake shares make's name but not its syntax, so running it against a
// GNU Makefile would fail confusingly. It must never be a candidate.
func TestMakeDoesNotResolveToNmake(t *testing.T) {
	for _, tool := range externalTools {
		if tool.name != "make" {
			continue
		}
		for _, c := range tool.candidates {
			if strings.EqualFold(c.File, "nmake.exe") {
				t.Error("make must not fall back to nmake.exe, which is not GNU make")
			}
		}
	}
}
