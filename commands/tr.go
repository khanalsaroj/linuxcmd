package commands

import (
	"bufio"
	"fmt"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type trCommand struct{}

func (trCommand) Name() string    { return "tr" }
func (trCommand) Summary() string { return "translate or delete characters" }

var trSpec = parser.Spec{
	{Short: 'd', Long: "delete"},
	{Short: 's', Long: "squeeze-repeats"},
}

// expandSet turns a tr character-set spec into its literal runes,
// expanding "a-z"-style ranges. Backslash escapes and POSIX classes like
// [:alpha:] are not supported.
func expandSet(spec string) []rune {
	runes := []rune(spec)
	var out []rune
	for i := 0; i < len(runes); i++ {
		if i+2 < len(runes) && runes[i+1] == '-' && runes[i] <= runes[i+2] {
			for c := runes[i]; c <= runes[i+2]; c++ {
				out = append(out, c)
			}
			i += 2
			continue
		}
		out = append(out, runes[i])
	}
	return out
}

func runeSet(runes []rune) map[rune]bool {
	m := make(map[rune]bool, len(runes))
	for _, r := range runes {
		m[r] = true
	}
	return m
}

func (trCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, trSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tr: %s\n", err)
		return command.ExitUsage
	}

	deleteMode := res.Bool('d', "delete")
	squeeze := res.Bool('s', "squeeze-repeats")

	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: tr [-d] [-s] SET1 [SET2]")
		return command.ExitUsage
	}
	set1 := expandSet(res.Positional[0])

	var mapping map[rune]rune
	var deleteSet map[rune]bool
	var squeezeSet map[rune]bool

	switch {
	case deleteMode:
		deleteSet = runeSet(set1)
		if squeeze && len(res.Positional) >= 2 {
			squeezeSet = runeSet(expandSet(res.Positional[1]))
		}
	case len(res.Positional) >= 2:
		set2 := expandSet(res.Positional[1])
		if len(set2) == 0 {
			fmt.Fprintln(ctx.Stderr, "tr: SET2 must not be empty")
			return command.ExitUsage
		}
		mapping = make(map[rune]rune, len(set1))
		for i, c := range set1 {
			if i < len(set2) {
				mapping[c] = set2[i]
			} else {
				mapping[c] = set2[len(set2)-1]
			}
		}
		if squeeze {
			squeezeSet = runeSet(set2)
		}
	case squeeze:
		squeezeSet = runeSet(set1)
	default:
		fmt.Fprintln(ctx.Stderr, "tr: missing SET2 (or use -d/-s)")
		return command.ExitUsage
	}

	reader := bufio.NewReader(ctx.Stdin)
	writer := bufio.NewWriter(ctx.Stdout)

	var lastOut rune
	haveLast := false
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		out := r
		skip := false
		if deleteMode {
			if deleteSet[r] {
				skip = true
			}
		} else if mapping != nil {
			if mapped, ok := mapping[r]; ok {
				out = mapped
			}
		}
		if skip {
			continue
		}
		if squeeze && squeezeSet[out] && haveLast && lastOut == out {
			continue
		}
		writer.WriteRune(out)
		lastOut = out
		haveLast = true
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(ctx.Stderr, "tr: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(trCommand{}) }
