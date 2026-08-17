package commands

import (
	"fmt"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
)

type printfCommand struct{}

func (printfCommand) Name() string    { return "printf" }
func (printfCommand) Summary() string { return "format and print arguments" }

func (printfCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: printf FORMAT [ARGUMENT]...")
		return command.ExitUsage
	}
	format := unescapeFormat(ctx.Args[0])
	values := ctx.Args[1:]

	out, err := renderPrintf(format, values)
	fmt.Fprint(ctx.Stdout, out)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "printf: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// unescapeFormat expands the small set of backslash escapes printf's
// format string supports (\n, \t, \\, \", \\\\).
func unescapeFormat(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// renderPrintf applies a subset of printf's %s/%d/%f/%x/%% conversions,
// cycling the format string over the argument list if there are more
// arguments than conversions (as POSIX printf does).
func renderPrintf(format string, values []string) (string, error) {
	var out strings.Builder
	vi := 0
	next := func() string {
		if vi < len(values) {
			v := values[vi]
			vi++
			return v
		}
		return ""
	}

	apply := func(f string) error {
		for i := 0; i < len(f); i++ {
			if f[i] != '%' || i+1 >= len(f) {
				out.WriteByte(f[i])
				continue
			}
			i++
			switch f[i] {
			case '%':
				out.WriteByte('%')
			case 's':
				out.WriteString(next())
			case 'd', 'i':
				n, err := strconv.ParseInt(strings.TrimSpace(next()), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer argument")
				}
				fmt.Fprintf(&out, "%d", n)
			case 'f':
				n, err := strconv.ParseFloat(strings.TrimSpace(next()), 64)
				if err != nil {
					return fmt.Errorf("invalid float argument")
				}
				fmt.Fprintf(&out, "%f", n)
			case 'x':
				n, err := strconv.ParseInt(strings.TrimSpace(next()), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer argument")
				}
				fmt.Fprintf(&out, "%x", n)
			default:
				out.WriteByte('%')
				out.WriteByte(f[i])
			}
		}
		return nil
	}

	if len(values) == 0 {
		err := apply(format)
		return out.String(), err
	}
	for vi < len(values) {
		before := vi
		if err := apply(format); err != nil {
			return out.String(), err
		}
		if vi == before {
			// Format string has no conversions; avoid looping forever.
			break
		}
	}
	return out.String(), nil
}

func init() { command.Register(printfCommand{}) }
