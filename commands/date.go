package commands

import (
	"fmt"
	"strings"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type dateCommand struct{}

func (dateCommand) Name() string    { return "date" }
func (dateCommand) Summary() string { return "print or format the current date and time" }

var dateSpec = parser.Spec{
	{Short: 'u', Long: "utc"},
}

func (dateCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, dateSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "date: %s\n", err)
		return command.ExitUsage
	}

	now := time.Now()
	if res.Bool('u', "utc") {
		now = now.UTC()
	}

	format := "+%a %b %e %H:%M:%S %Z %Y"
	for _, a := range res.Positional {
		if strings.HasPrefix(a, "+") {
			format = a
			break
		}
	}

	fmt.Fprintln(ctx.Stdout, formatDate(now, format[1:]))
	return command.ExitSuccess
}

// formatDate maps a subset of strftime-style "%" tokens (as accepted by
// GNU date's +FORMAT) onto Go's reference-time layout.
func formatDate(t time.Time, spec string) string {
	var b strings.Builder
	for i := 0; i < len(spec); i++ {
		if spec[i] != '%' || i+1 >= len(spec) {
			b.WriteByte(spec[i])
			continue
		}
		i++
		switch spec[i] {
		case 'Y':
			b.WriteString(t.Format("2006"))
		case 'y':
			b.WriteString(t.Format("06"))
		case 'm':
			b.WriteString(t.Format("01"))
		case 'd':
			b.WriteString(t.Format("02"))
		case 'e':
			b.WriteString(fmt.Sprintf("%2d", t.Day()))
		case 'H':
			b.WriteString(t.Format("15"))
		case 'I':
			b.WriteString(t.Format("03"))
		case 'M':
			b.WriteString(t.Format("04"))
		case 'S':
			b.WriteString(t.Format("05"))
		case 'p':
			b.WriteString(t.Format("PM"))
		case 'a':
			b.WriteString(t.Format("Mon"))
		case 'A':
			b.WriteString(t.Format("Monday"))
		case 'b', 'h':
			b.WriteString(t.Format("Jan"))
		case 'B':
			b.WriteString(t.Format("January"))
		case 'Z':
			b.WriteString(t.Format("MST"))
		case 'z':
			b.WriteString(t.Format("-0700"))
		case 'j':
			b.WriteString(fmt.Sprintf("%03d", t.YearDay()))
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case '%':
			b.WriteByte('%')
		case 's':
			b.WriteString(fmt.Sprintf("%d", t.Unix()))
		default:
			b.WriteByte('%')
			b.WriteByte(spec[i])
		}
	}
	return b.String()
}

func init() { command.Register(dateCommand{}) }
