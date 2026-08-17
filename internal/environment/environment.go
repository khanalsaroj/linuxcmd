// Package environment maps Linux-style environment variable names ($HOME,
// $USER, ...) onto their Windows equivalents and expands $VAR / ${VAR}
// references in strings, the way a POSIX shell would before a command
// ever sees its arguments.
package environment

import (
	"os"
	"regexp"
)

// aliases maps a Linux-conventional variable name to the Windows
// environment variable that holds the equivalent value. Names not listed
// here fall back to a direct os.Getenv lookup, so existing Windows
// variables like %USERNAME% remain accessible as $USERNAME too.
var aliases = map[string]string{
	"HOME":   "USERPROFILE",
	"USER":   "USERNAME",
	"TMP":    "TEMP",
	"TMPDIR": "TEMP",
	"SHELL":  "COMSPEC",
	"HOSTNAME": "COMPUTERNAME",
}

// Lookup resolves a Linux-style variable name to its value, following the
// alias table above. PWD is computed dynamically since Windows does not
// keep it as an environment variable.
func Lookup(name string) (string, bool) {
	if name == "PWD" {
		if wd, err := os.Getwd(); err == nil {
			return wd, true
		}
		return "", false
	}
	if winName, ok := aliases[name]; ok {
		if v, ok := os.LookupEnv(winName); ok {
			return v, true
		}
	}
	return os.LookupEnv(name)
}

// varPattern matches $VAR and ${VAR} references.
var varPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// Expand replaces $VAR / ${VAR} references in s with their resolved
// values. Unknown variables expand to an empty string, matching typical
// shell behavior.
func Expand(s string) string {
	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := varPattern.FindStringSubmatch(match)
		name := sub[1]
		if name == "" {
			name = sub[2]
		}
		v, _ := Lookup(name)
		return v
	})
}
