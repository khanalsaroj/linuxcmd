package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"linuxcmd/internal/command"
)

// run executes a registered command by name with the given args, in the
// current working directory, and returns its exit code plus captured
// stdout/stderr. Tests use command.Lookup (the same path the real
// linuxcmd.exe entrypoint uses) rather than constructing the unexported
// command structs directly, so they exercise the actual registration.
func run(t *testing.T, name string, args ...string) (exitCode int, stdout, stderr string) {
	t.Helper()
	cmd, ok := command.Lookup(name)
	if !ok {
		t.Fatalf("command %q is not registered", name)
	}
	var outBuf, errBuf bytes.Buffer
	ctx := &command.Context{
		CommandName: name,
		Args:        args,
		Stdin:       bytes.NewReader(nil),
		Stdout:      &outBuf,
		Stderr:      &errBuf,
	}
	exitCode = cmd.Run(ctx)
	return exitCode, outBuf.String(), errBuf.String()
}

// runIn is like run but first switches the process working directory to
// dir, restoring the original on test cleanup. Needed for commands (pwd,
// relative-path resolution) whose behavior depends on cwd.
func runIn(t *testing.T, dir, name string, args ...string) (exitCode int, stdout, stderr string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return run(t, name, args...)
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// canonicalPath resolves path the way realpath (and filepath.EvalSymlinks)
// does. On Windows this also expands 8.3 short names, which matters on CI
// runners where TEMP is something like C:\Users\RUNNER~1\AppData\Local\Temp
// and t.TempDir() therefore hands back the short form.
func canonicalPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}
