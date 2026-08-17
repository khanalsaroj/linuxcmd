package commands

import (
	"fmt"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/environment"
	"linuxcmd/internal/parser"
)

type echoCommand struct{}

func (echoCommand) Name() string    { return "echo" }
func (echoCommand) Summary() string { return "display a line of text" }

var echoSpec = parser.Spec{
	{Short: 'n'},
	{Short: 'e'},
}

func (echoCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, echoSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "echo: %s\n", err)
		return command.ExitUsage
	}

	noNewline := res.Bool('n', "")
	escapes := res.Bool('e', "")

	parts := make([]string, len(res.Positional))
	for i, a := range res.Positional {
		expanded := environment.Expand(a)
		if escapes {
			expanded = interpretEscapes(expanded)
		}
		parts[i] = expanded
	}

	out := strings.Join(parts, " ")
	if noNewline {
		fmt.Fprint(ctx.Stdout, out)
	} else {
		fmt.Fprintln(ctx.Stdout, out)
	}
	return command.ExitSuccess
}

// interpretEscapes expands the small set of backslash escapes GNU echo
// -e supports.
func interpretEscapes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		default:
			b.WriteByte(s[i])
			b.WriteByte(s[i+1])
		}
		i++
	}
	return b.String()
}

func init() { command.Register(echoCommand{}) }
