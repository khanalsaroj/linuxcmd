// Package parser implements a small POSIX-style option parser. The
// standard library's flag package doesn't support combined short flags
// (e.g. "-la" meaning "-l -a") or GNU-style "--long" options, both of
// which Linux users expect, so commands describe their accepted options
// as a Spec and call Parse instead of splitting strings by hand.
//
// Note that shell-style tokenization (splitting a command line on spaces,
// honoring quotes) is NOT this package's job: Windows' own process
// creation already splits os.Args into properly quoted tokens before a
// command ever sees them, so each element of args is already one token.
package parser

import (
	"fmt"
)

// Option describes one flag a command accepts.
type Option struct {
	Short  byte   // short flag letter, e.g. 'l'; 0 if none
	Long   string // long flag name, e.g. "all"; "" if none
	HasArg bool   // whether the flag takes a value
}

// Result holds the outcome of parsing: which flags were set, their
// values, and the remaining positional (non-flag) arguments.
type Result struct {
	bools      map[byte]bool
	longBools  map[string]bool
	values     map[byte]string
	longValues map[string]string
	Positional []string
}

// Bool reports whether a boolean flag was set, checked by both its short
// and long name (either may be zero-valued if the flag has no alias).
func (r *Result) Bool(short byte, long string) bool {
	if short != 0 && r.bools[short] {
		return true
	}
	if long != "" && r.longBools[long] {
		return true
	}
	return false
}

// Value returns a flag's argument and whether it was provided.
func (r *Result) Value(short byte, long string) (string, bool) {
	if short != 0 {
		if v, ok := r.values[short]; ok {
			return v, true
		}
	}
	if long != "" {
		if v, ok := r.longValues[long]; ok {
			return v, true
		}
	}
	return "", false
}

// Spec is the set of options a command recognizes.
type Spec []Option

func (s Spec) findShort(c byte) (Option, bool) {
	for _, o := range s {
		if o.Short == c {
			return o, true
		}
	}
	return Option{}, false
}

func (s Spec) findLong(name string) (Option, bool) {
	for _, o := range s {
		if o.Long == name {
			return o, true
		}
	}
	return Option{}, false
}

// Parse splits args into recognized flags and positional arguments per
// spec. "--" stops flag parsing; everything after it is positional.
// A single "-" is treated as positional (conventionally meaning stdin).
func Parse(args []string, spec Spec) (*Result, error) {
	res := &Result{
		bools:      map[byte]bool{},
		longBools:  map[string]bool{},
		values:     map[byte]string{},
		longValues: map[string]string{},
	}

	noMoreFlags := false
	for i := 0; i < len(args); i++ {
		a := args[i]

		if noMoreFlags || a == "-" || len(a) == 0 || a[0] != '-' {
			res.Positional = append(res.Positional, a)
			continue
		}
		if a == "--" {
			noMoreFlags = true
			continue
		}

		if len(a) >= 2 && a[1] == '-' {
			// Long option: --name or --name=value
			name := a[2:]
			value := ""
			hasValue := false
			for j := 0; j < len(name); j++ {
				if name[j] == '=' {
					value = name[j+1:]
					name = name[:j]
					hasValue = true
					break
				}
			}
			opt, ok := spec.findLong(name)
			if !ok {
				return nil, fmt.Errorf("unrecognized option '--%s'", name)
			}
			if opt.HasArg {
				if !hasValue {
					if i+1 >= len(args) {
						return nil, fmt.Errorf("option '--%s' requires an argument", name)
					}
					i++
					value = args[i]
				}
				res.longValues[name] = value
			} else {
				res.longBools[name] = true
			}
			continue
		}

		// Short option cluster: -la, -n5, -n 5
		chars := a[1:]
		for j := 0; j < len(chars); j++ {
			c := chars[j]
			opt, ok := spec.findShort(c)
			if !ok {
				return nil, fmt.Errorf("invalid option -- '%c'", c)
			}
			if !opt.HasArg {
				res.bools[c] = true
				continue
			}
			// Remainder of this token (if any) is the value; otherwise
			// consume the next argument.
			if j+1 < len(chars) {
				res.values[c] = chars[j+1:]
			} else {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("option requires an argument -- '%c'", c)
				}
				i++
				res.values[c] = args[i]
			}
			break
		}
	}

	return res, nil
}
