package commands

import (
	"fmt"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// blkidCommand reports volume identifiers. The Linux fields map onto
// Windows ones closely enough to be useful, with one caveat worth
// knowing: Windows' volume serial number is a 32-bit value assigned at
// format time, not a 128-bit filesystem UUID, so it is shorter and it
// changes whenever the volume is reformatted.
type blkidCommand struct{}

func (blkidCommand) Name() string    { return "blkid" }
func (blkidCommand) Summary() string { return "show volume identifiers and filesystem types" }

var blkidSpec = parser.Spec{
	{Short: 'o', HasArg: true}, // output format: full (default) or value
	{Short: 's', HasArg: true}, // restrict to a single tag
}

func (blkidCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, blkidSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "blkid: %s\n", err)
		return command.ExitUsage
	}

	vols, err := enumerateVolumes()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "blkid: %s\n", err)
		return command.ExitFailure
	}

	// Operands select specific devices, accepting either "C:" or "C:\".
	wanted := map[string]bool{}
	for _, arg := range res.Positional {
		wanted[strings.ToUpper(strings.TrimSuffix(arg, `\`))] = true
	}

	onlyTag, hasTag := res.Value('s', "")
	valuesOnly := false
	if v, ok := res.Value('o', ""); ok {
		valuesOnly = v == "value"
	}

	matched := false
	for _, v := range vols {
		if len(wanted) > 0 && !wanted[strings.ToUpper(v.Letter)] {
			continue
		}
		// A volume with no filesystem (an empty optical drive) has
		// nothing to report.
		if v.FSType == "" {
			continue
		}
		matched = true

		tags := [][2]string{}
		if v.Label != "" {
			tags = append(tags, [2]string{"LABEL", v.Label})
		}
		tags = append(tags, [2]string{"UUID", formatVolumeSerial(v.Serial)})
		tags = append(tags, [2]string{"TYPE", v.FSType})

		var parts []string
		for _, t := range tags {
			if hasTag && !strings.EqualFold(t[0], onlyTag) {
				continue
			}
			if valuesOnly {
				parts = append(parts, t[1])
			} else {
				parts = append(parts, fmt.Sprintf(`%s="%s"`, t[0], t[1]))
			}
		}
		if len(parts) == 0 {
			continue
		}

		if valuesOnly {
			fmt.Fprintln(ctx.Stdout, strings.Join(parts, "\n"))
		} else {
			// blkid separates the device from its tags with ": ". A
			// Windows drive letter already ends in a colon, so printing
			// both would give "C:: UUID=..."; the letter's own colon
			// serves as the separator instead.
			fmt.Fprintf(ctx.Stdout, "%s %s\n", v.Letter, strings.Join(parts, " "))
		}
	}

	// blkid exits 2 when nothing matched, which scripts rely on.
	if !matched {
		return 2
	}
	return command.ExitSuccess
}

// formatVolumeSerial renders a Windows volume serial the way Windows
// itself displays it, as two hex groups (1A2B-3C4D).
func formatVolumeSerial(serial uint32) string {
	return fmt.Sprintf("%04X-%04X", serial>>16, serial&0xFFFF)
}

func init() { command.Register(blkidCommand{}) }
