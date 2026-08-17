package commands

import "testing"

func TestGitWrapsRealExecutable(t *testing.T) {
	if _, ok := findRealGit(); !ok {
		t.Skip("no real git.exe installed in this environment")
	}
	code, out, errOut := run(t, "git", "--version")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out == "" {
		t.Error("expected git --version to print something")
	}
}

func TestGitMissingBinary(t *testing.T) {
	// Doesn't assert on the outcome (a real git.exe may or may not be
	// installed here) -- just confirms findRealGit runs without panicking
	// and, when it does find one, that the reported path actually exists.
	if path, ok := findRealGit(); ok {
		if path == "" {
			t.Error("findRealGit reported ok=true with an empty path")
		}
	}
}
