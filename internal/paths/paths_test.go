package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeHome(t *testing.T) {
	home := Home()
	got := Normalize("~")
	if got != home {
		t.Errorf("Normalize(~) = %q, want %q", got, home)
	}
	got = Normalize("~/Documents")
	want := filepath.Join(home, "Documents")
	if got != want {
		t.Errorf("Normalize(~/Documents) = %q, want %q", got, want)
	}
}

func TestNormalizeTmp(t *testing.T) {
	got := Normalize("/tmp")
	want := os.TempDir()
	if got != want {
		t.Errorf("Normalize(/tmp) = %q, want %q", got, want)
	}
	got = Normalize("/tmp/foo.txt")
	want = filepath.Join(os.TempDir(), "foo.txt")
	if got != want {
		t.Errorf("Normalize(/tmp/foo.txt) = %q, want %q", got, want)
	}
}

func TestNormalizeDevNull(t *testing.T) {
	if got := Normalize("/dev/null"); got != os.DevNull {
		t.Errorf("Normalize(/dev/null) = %q, want %q", got, os.DevNull)
	}
}

func TestNormalizeDriveShorthand(t *testing.T) {
	got := Normalize("/c/Users/foo")
	want := `C:\Users\foo`
	if got != want {
		t.Errorf("Normalize(/c/Users/foo) = %q, want %q", got, want)
	}
}

func TestNormalizeWindowsAbsolute(t *testing.T) {
	cases := []string{`C:\Users\foo`, "C:/Users/foo"}
	for _, c := range cases {
		got := Normalize(c)
		want := `C:\Users\foo`
		if got != want {
			t.Errorf("Normalize(%q) = %q, want %q", c, got, want)
		}
	}
}

func TestNormalizeBareLinuxRoot(t *testing.T) {
	got := Normalize("/etc")
	if !strings.HasSuffix(got, `\etc`) {
		t.Errorf("Normalize(/etc) = %q, want it to end with \\etc", got)
	}
	if !strings.Contains(got, ":") {
		t.Errorf("Normalize(/etc) = %q, want a drive-qualified path", got)
	}
}

func TestNormalizeRelative(t *testing.T) {
	got := Normalize("foo/bar")
	want := filepath.FromSlash("foo/bar")
	if got != want {
		t.Errorf("Normalize(foo/bar) = %q, want %q", got, want)
	}
}

func TestResolveMakesAbsolute(t *testing.T) {
	got, err := Resolve("foo")
	if err != nil {
		t.Fatalf("Resolve(foo) error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("Resolve(foo) = %q, want an absolute path", got)
	}
}

func TestExpandGlobsMatchesExistingFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c.log"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pattern := filepath.Join(dir, "*.txt")
	got := ExpandGlobs([]string{pattern})
	if len(got) != 2 {
		t.Fatalf("ExpandGlobs(%q) = %v, want 2 matches", pattern, got)
	}
}

func TestExpandGlobsPassesThroughNonMatching(t *testing.T) {
	got := ExpandGlobs([]string{"no-such-glob-*-pattern-xyz"})
	if len(got) != 1 || got[0] != "no-such-glob-*-pattern-xyz" {
		t.Errorf("ExpandGlobs with no matches should pass the literal arg through, got %v", got)
	}
}

func TestExpandGlobsLeavesPlainArgsAlone(t *testing.T) {
	got := ExpandGlobs([]string{"plain.txt", "other.txt"})
	if len(got) != 2 || got[0] != "plain.txt" || got[1] != "other.txt" {
		t.Errorf("ExpandGlobs should leave non-glob args unchanged, got %v", got)
	}
}
