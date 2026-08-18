package commands

import (
	"path/filepath"
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

func TestNameiWalksEveryComponent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "file.txt")
	mustWriteFile(t, path, "x")

	code, out, errOut := run(t, "namei", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.HasPrefix(out, "f: ") {
		t.Errorf("namei output = %q, want an 'f: ' header line", out)
	}
	if !strings.Contains(out, " d deep") {
		t.Errorf("namei output = %q, want the intermediate directory marked 'd'", out)
	}
	if !strings.Contains(out, " - file.txt") {
		t.Errorf("namei output = %q, want the regular file marked '-'", out)
	}
}

// The first component printed is the resolved Windows volume, which is
// where the Linux-to-Windows path translation becomes visible.
func TestNameiShowsResolvedVolume(t *testing.T) {
	code, out, errOut := run(t, "namei", "~")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("namei ~ output = %q, want a header and at least one component", out)
	}
	if !strings.Contains(lines[1], ":\\") {
		t.Errorf("first component = %q, want a Windows volume root", lines[1])
	}
}

func TestNameiStopsAtMissingComponent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "absent", "deeper", "file.txt")

	code, out, _ := run(t, "namei", path)
	if code == 0 {
		t.Error("expected a nonzero exit when a component does not exist")
	}
	if !strings.Contains(out, "? absent") {
		t.Errorf("namei output = %q, want the missing component marked '?'", out)
	}
	// The header line echoes the whole argument, so only the component
	// lines below it are checked for having stopped early.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines[1:] {
		if strings.Contains(line, "deeper") {
			t.Errorf("namei walked past the first unresolvable component: %q", line)
		}
	}
	if !strings.HasSuffix(lines[len(lines)-1], "No such file or directory") {
		t.Errorf("last line = %q, want the walk to end on the error", lines[len(lines)-1])
	}
}

func TestNameiModesFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	mustWriteFile(t, path, "x")

	code, out, errOut := run(t, "namei", "-m", path)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "-rw") && !strings.Contains(out, "-r-") {
		t.Errorf("namei -m output = %q, want a mode string for the file", out)
	}
}

func TestNameiRequiresAnOperand(t *testing.T) {
	code, _, errOut := run(t, "namei")
	if code == 0 {
		t.Error("expected a nonzero exit with no operand")
	}
	if !strings.Contains(errOut, "usage:") {
		t.Errorf("stderr = %q, want a usage line", errOut)
	}
}

func TestManRendersCuratedPage(t *testing.T) {
	code, out, errOut := run(t, "man", "namei")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	for _, section := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "OPTIONS", "ON WINDOWS", "LINUXCMD"} {
		if !strings.Contains(out, section) {
			t.Errorf("man namei is missing the %s section:\n%s", section, out)
		}
	}
}

// A command with no hand-written page still gets one, built from the
// registry, rather than an error.
func TestManGeneratesPageForUndocumentedCommand(t *testing.T) {
	code, out, errOut := run(t, "man", "true")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "true") {
		t.Errorf("man true = %q, want a generated page", out)
	}
}

// Commands that diverge from Linux for a shared reason get the shared
// note, so the explanation cannot drift between them.
func TestManAppliesGroupNotes(t *testing.T) {
	code, out, errOut := run(t, "man", "kill")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "ON WINDOWS") || !strings.Contains(out, "signal") {
		t.Errorf("man kill = %q, want the shared signal note", out)
	}
}

func TestManApropos(t *testing.T) {
	code, out, errOut := run(t, "man", "-k", "clipboard")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "xclip") || !strings.Contains(out, "xsel") {
		t.Errorf("man -k clipboard = %q, want both clipboard commands", out)
	}
}

func TestManAproposReportsNoMatch(t *testing.T) {
	code, _, errOut := run(t, "man", "-k", "zzzznotathing")
	if code == 0 {
		t.Error("expected a nonzero exit when nothing matches")
	}
	if !strings.Contains(errOut, "nothing appropriate") {
		t.Errorf("stderr = %q, want man's no-match message", errOut)
	}
}

func TestManWhatis(t *testing.T) {
	code, out, errOut := run(t, "man", "-f", "od")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.HasPrefix(out, "od (1) - ") {
		t.Errorf("man -f od = %q, want a whatis line", out)
	}
}

func TestManUnknownTopic(t *testing.T) {
	code, _, errOut := run(t, "man", "definitelynotacommand")
	if code == 0 {
		t.Error("expected a nonzero exit for an unknown topic")
	}
	if !strings.Contains(errOut, "No manual entry") {
		t.Errorf("stderr = %q, want man's unknown-entry message", errOut)
	}
}

// A curated page for a command that no longer exists would be dead
// documentation nobody could reach.
func TestManPagesAllNameRegisteredCommands(t *testing.T) {
	for _, name := range manPageNames() {
		if _, ok := command.Lookup(name); !ok {
			t.Errorf("manPages has an entry for %q, which is not a registered command", name)
		}
	}
}

func TestManGroupsAllNameRegisteredCommands(t *testing.T) {
	for _, g := range manGroups {
		for _, name := range g.commands {
			if _, ok := command.Lookup(name); !ok {
				t.Errorf("manGroups references %q, which is not a registered command", name)
			}
		}
	}
}
