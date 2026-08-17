// Package tests contains integration tests that build the real
// linuxcmd.exe and invoke it as a subprocess, the way a user's CMD
// window actually would, rather than calling command.Run in-process
// (which the unit tests in commands/ already cover).
package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildEngine compiles linuxcmd.exe once for the whole test binary run
// and returns its path.
func buildEngine(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	exePath := filepath.Join(dir, "linuxcmd.exe")

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "build", "-o", exePath, "./cmd/linuxcmd")
	cmd.Dir = repoRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build linuxcmd.exe: %v\n%s", err, stderr.String())
	}
	return exePath
}

// invoke runs the built engine using the "linuxcmd <command>" subcommand
// form (no per-command launcher needed) with cwd set to dir.
func invoke(t *testing.T, enginePath, dir, name string, args ...string) (int, string, string) {
	t.Helper()
	full := append([]string{name}, args...)
	cmd := exec.Command(enginePath, full...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run %s: %v", name, err)
		}
	}
	return code, stdout.String(), stderr.String()
}

func TestIntegrationEndToEndWorkflow(t *testing.T) {
	engine := buildEngine(t)
	dir := t.TempDir()

	// mkdir, then verify with ls
	if code, _, errOut := invoke(t, engine, dir, "mkdir", "project"); code != 0 {
		t.Fatalf("mkdir failed: %s", errOut)
	}
	if code, out, errOut := invoke(t, engine, dir, "ls", dir); code != 0 || !strings.Contains(out, "project") {
		t.Fatalf("ls did not show project dir: out=%q err=%q code=%d", out, errOut, code)
	}

	// touch + cat round trip
	target := filepath.Join(dir, "project", "hello.txt")
	if code, _, errOut := invoke(t, engine, dir, "touch", target); code != 0 {
		t.Fatalf("touch failed: %s", errOut)
	}
	if err := os.WriteFile(target, []byte("hello from integration test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if code, out, errOut := invoke(t, engine, dir, "cat", target); code != 0 || out != "hello from integration test\n" {
		t.Fatalf("cat mismatch: out=%q err=%q code=%d", out, errOut, code)
	}

	// cp + mv + rm
	copyTarget := filepath.Join(dir, "project", "copy.txt")
	if code, _, errOut := invoke(t, engine, dir, "cp", target, copyTarget); code != 0 {
		t.Fatalf("cp failed: %s", errOut)
	}
	renamed := filepath.Join(dir, "project", "renamed.txt")
	if code, _, errOut := invoke(t, engine, dir, "mv", copyTarget, renamed); code != 0 {
		t.Fatalf("mv failed: %s", errOut)
	}
	if code, _, errOut := invoke(t, engine, dir, "rm", renamed); code != 0 {
		t.Fatalf("rm failed: %s", errOut)
	}
	if _, err := os.Stat(renamed); !os.IsNotExist(err) {
		t.Error("expected renamed.txt to be removed")
	}

	// pwd reflects cwd
	if code, out, errOut := invoke(t, engine, dir, "pwd"); code != 0 || !strings.EqualFold(strings.TrimSpace(out), dir) {
		t.Fatalf("pwd mismatch: out=%q err=%q code=%d, want %q", out, errOut, code, dir)
	}

	// grep finds the line we wrote
	if code, out, errOut := invoke(t, engine, dir, "grep", "integration", target); code != 0 || !strings.Contains(out, "integration") {
		t.Fatalf("grep mismatch: out=%q err=%q code=%d", out, errOut, code)
	}

	// exit code propagation for a real failure
	if code, _, _ := invoke(t, engine, dir, "cat", filepath.Join(dir, "does-not-exist.txt")); code == 0 {
		t.Error("expected nonzero exit code for cat on a missing file")
	}

	// rmdir cleans up the now-empty directory
	if code, _, errOut := invoke(t, engine, dir, "rm", target); code != 0 {
		t.Fatalf("rm failed: %s", errOut)
	}
	if code, _, errOut := invoke(t, engine, dir, "rmdir", filepath.Join(dir, "project")); code != 0 {
		t.Fatalf("rmdir failed: %s", errOut)
	}
}

func TestIntegrationCommandNotFound(t *testing.T) {
	engine := buildEngine(t)
	dir := t.TempDir()
	code, _, errOut := invoke(t, engine, dir, "not-a-real-command")
	if code != 127 {
		t.Errorf("exit code = %d, want 127 for an unknown command", code)
	}
	if !strings.Contains(errOut, "command not found") {
		t.Errorf("expected 'command not found' message, got %q", errOut)
	}
}

func TestIntegrationPerCommandLauncherDispatch(t *testing.T) {
	engine := buildEngine(t)
	dir := t.TempDir()

	// Create a hardlink named pwd.exe next to linuxcmd.exe, the same way
	// scripts/build.ps1 and installer/install.ps1 do, and confirm argv[0]
	// dispatch (rather than the "linuxcmd <cmd>" subcommand form) works.
	launcher := filepath.Join(filepath.Dir(engine), "pwd.exe")
	if err := os.Link(engine, launcher); err != nil {
		t.Skipf("hardlinks not supported in this environment: %v", err)
	}

	cmd := exec.Command(launcher)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("pwd.exe failed: %v (%s)", err, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); !strings.EqualFold(got, dir) {
		t.Errorf("pwd.exe output = %q, want %q", got, dir)
	}
}
