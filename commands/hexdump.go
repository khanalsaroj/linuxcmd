package commands

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// hexdumpCommand implements the common hexdump views. It overlaps with
// xxd and od on purpose: all three are reached for by muscle memory and
// they disagree on defaults, so having only one of them means output
// that does not match what a user pasted from a Linux box. The custom
// format-string language (hexdump -e) is not implemented; -C, -b, -c,
// -d, -o and -x cover what hexdump is actually used for interactively.
type hexdumpCommand struct{}

func (hexdumpCommand) Name() string    { return "hexdump" }
func (hexdumpCommand) Summary() string { return "display file contents in hex, decimal, or octal" }

var hexdumpSpec = parser.Spec{
	{Short: 'C', Long: "canonical"},
	{Short: 'b', Long: "one-byte-octal"},
	{Short: 'c', Long: "one-byte-char"},
	{Short: 'd', Long: "two-bytes-decimal"},
	{Short: 'o', Long: "two-bytes-octal"},
	{Short: 'x', Long: "two-bytes-hex"},
	{Short: 'n', Long: "length", HasArg: true},
	{Short: 's', Long: "skip", HasArg: true},
	{Short: 'v', Long: "no-squeezing"},
}

func (hexdumpCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, hexdumpSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "hexdump: %s\n", err)
		return command.ExitUsage
	}

	skip := int64(0)
	if v, ok := res.Value('s', "skip"); ok {
		skip, err = parseByteCount(v)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "hexdump: invalid skip value: %s\n", err)
			return command.ExitUsage
		}
	}
	length, hasLength := int64(0), false
	if v, ok := res.Value('n', "length"); ok {
		length, err = parseByteCount(v)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "hexdump: invalid length value: %s\n", err)
			return command.ExitUsage
		}
		hasLength = true
	}

	var render func(chunk []byte, offset int64) string
	switch {
	case res.Bool('C', "canonical"):
		render = hexdumpCanonicalLine
	case res.Bool('b', "one-byte-octal"):
		render = hexdumpByteLine(3, func(b byte) string { return fmt.Sprintf("%03o", b) })
	case res.Bool('c', "one-byte-char"):
		render = hexdumpByteLine(3, hexdumpChar)
	case res.Bool('d', "two-bytes-decimal"):
		render = hexdumpWordLine(func(v uint16) string { return fmt.Sprintf("%05d", v) })
	case res.Bool('o', "two-bytes-octal"):
		render = hexdumpWordLine(func(v uint16) string { return fmt.Sprintf("%06o", v) })
	default:
		// hexdump's default and -x are both the two-byte hex view.
		render = hexdumpWordLine(func(v uint16) string { return fmt.Sprintf("%04x", v) })
	}

	in, err := openDumpInput(ctx, "hexdump", paths.ExpandGlobs(res.Positional))
	if err != nil {
		return command.ExitFailure
	}
	defer in.Close()

	if err := in.limit(skip, length, hasLength); err != nil {
		fmt.Fprintf(ctx.Stderr, "hexdump: %s\n", err)
		return command.ExitFailure
	}

	if err := hexdumpRun(ctx, in.reader, render, skip, res.Bool('v', "no-squeezing")); err != nil {
		fmt.Fprintf(ctx.Stderr, "hexdump: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func hexdumpRun(ctx *command.Context, r io.Reader, render func([]byte, int64) string, startOffset int64, verbose bool) error {
	buf := make([]byte, 16)
	offset := startOffset
	dw := newDumpWriter(ctx.Stdout, verbose)

	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			chunk := buf[:n]
			dw.row(chunk, render(chunk, offset))
			offset += int64(n)
		}
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
	}
	// hexdump closes with the final offset, with no trailing data.
	fmt.Fprintf(ctx.Stdout, "%08x\n", offset)
	return nil
}

// hexdumpCanonicalLine renders "hexdump -C": offset, sixteen hex bytes
// split into two groups of eight, then the printable gutter.
func hexdumpCanonicalLine(chunk []byte, offset int64) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%08x ", offset)
	for i := 0; i < 16; i++ {
		if i == 8 {
			sb.WriteByte(' ')
		}
		if i < len(chunk) {
			fmt.Fprintf(&sb, " %02x", chunk[i])
		} else {
			sb.WriteString("   ")
		}
	}
	sb.WriteString("  |")
	for _, c := range chunk {
		if c >= 32 && c < 127 {
			sb.WriteByte(c)
		} else {
			sb.WriteByte('.')
		}
	}
	sb.WriteByte('|')
	return sb.String()
}

// hexdumpByteLine builds a renderer for the single-byte views (-b, -c),
// which print sixteen fixed-width fields per line.
func hexdumpByteLine(width int, format func(byte) string) func([]byte, int64) string {
	return func(chunk []byte, offset int64) string {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%08x", offset)
		for _, c := range chunk {
			fmt.Fprintf(&sb, " %*s", width, format(c))
		}
		return sb.String()
	}
}

// hexdumpWordLine builds a renderer for the two-byte views (-d, -o, -x),
// which read little-endian 16-bit words, eight to a line.
func hexdumpWordLine(format func(uint16) string) func([]byte, int64) string {
	return func(chunk []byte, offset int64) string {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%08x", offset)
		for i := 0; i < len(chunk); i += 2 {
			var v uint16
			if i+1 < len(chunk) {
				v = binary.LittleEndian.Uint16(chunk[i : i+2])
			} else {
				// A trailing odd byte is zero-extended, matching hexdump.
				v = uint16(chunk[i])
			}
			sb.WriteByte(' ')
			sb.WriteString(format(v))
		}
		return sb.String()
	}
}

// hexdumpChar renders hexdump -c, which uses C escapes where they exist
// and three-digit octal otherwise.
func hexdumpChar(c byte) string {
	if esc, ok := odEscapes[c]; ok {
		return esc
	}
	if c >= 32 && c < 127 {
		return string(c)
	}
	return fmt.Sprintf("%03o", c)
}

func init() { command.Register(hexdumpCommand{}) }
