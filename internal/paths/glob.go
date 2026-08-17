package paths

import (
	"path/filepath"
	"strings"
)

// ExpandGlobs expands shell-wildcard arguments (*, ?, [...]) into the
// paths they match. This exists because, unlike bash, cmd.exe never
// expands wildcards before invoking a program — each command receives
// the literal string "*.txt" and Windows programs are expected to expand
// it themselves. Without this, "rm *.txt" or "ls *.go" would silently do
// nothing useful, breaking the Linux-shell muscle memory this project is
// built around.
//
// An argument with no metacharacters, or one that matches nothing, is
// passed through unchanged so normal "No such file or directory"
// handling still applies to typos and missing files.
func ExpandGlobs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if !strings.ContainsAny(a, "*?[") {
			out = append(out, a)
			continue
		}
		matches, err := filepath.Glob(Normalize(a))
		if err != nil || len(matches) == 0 {
			out = append(out, a)
			continue
		}
		out = append(out, matches...)
	}
	return out
}
