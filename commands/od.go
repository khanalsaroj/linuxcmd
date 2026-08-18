package commands

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// odCommand implements GNU od's octal/typed dump. It complements xxd
// (which is hex-only and reversible) by covering the octal, decimal,
// character and floating-point views od is normally reached for, and it
// reads multi-byte values little-endian to match the host architecture.
type odCommand struct{}

func (odCommand) Name() string { return "od" }
func (odCommand) Summary() string {
	return "dump files in octal, hex, decimal, or character form"
}

var odSpec = parser.Spec{
	{Short: 'A', HasArg: true},      // offset radix: d, o, x, n
	{Short: 't', HasArg: true},      // output type: a, c, d, f, o, u, x with size
	{Short: 'j', HasArg: true},      // skip bytes
	{Short: 'N', HasArg: true},      // limit bytes
	{Short: 'v', Long: "verbose"},   // do not collapse duplicate lines
	{Short: 'a'},                    // named characters == -t a
	{Short: 'b'},                    // octal bytes == -t o1
	{Short: 'c'},                    // printable characters == -t c
	{Short: 'd'},                    // unsigned decimal shorts == -t u2
	{Short: 'o'},                    // octal shorts == -t o2
	{Short: 'x'},                    // hex shorts == -t x2
	{Short: 's'},                    // signed decimal shorts == -t d2
	{Short: 'i'},                    // signed decimal ints == -t d4
	{Short: 'l'},                    // signed decimal longs == -t d8
	{Long: "skip-bytes", HasArg: true},
	{Long: "read-bytes", HasArg: true},
	{Long: "address-radix", HasArg: true},
	{Long: "format", HasArg: true},
}

// odFormat describes one rendering of the byte stream: how many bytes
// make up an item, how many items sit on a line, and how an item prints.
type odFormat struct {
	size    int
	perLine int
	render  func(b []byte) string
}

func (odCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, odSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "od: %s\n", err)
		return command.ExitUsage
	}

	radix := "o"
	if v, ok := res.Value('A', "address-radix"); ok {
		radix = v
	}
	switch radix {
	case "d", "o", "x", "n":
	default:
		fmt.Fprintf(ctx.Stderr, "od: invalid output address radix '%s'; it must be one character from [doxn]\n", radix)
		return command.ExitUsage
	}

	typeSpec := ""
	switch {
	case res.Bool('a', ""):
		typeSpec = "a"
	case res.Bool('b', ""):
		typeSpec = "o1"
	case res.Bool('c', ""):
		typeSpec = "c"
	case res.Bool('d', ""):
		typeSpec = "u2"
	case res.Bool('o', ""):
		typeSpec = "o2"
	case res.Bool('x', ""):
		typeSpec = "x2"
	case res.Bool('s', ""):
		typeSpec = "d2"
	case res.Bool('i', ""):
		typeSpec = "d4"
	case res.Bool('l', ""):
		typeSpec = "d8"
	}
	if v, ok := res.Value('t', "format"); ok {
		typeSpec = v
	}
	if typeSpec == "" {
		typeSpec = "o2" // od's default view
	}

	format, err := parseODFormat(typeSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "od: %s\n", err)
		return command.ExitUsage
	}

	skip, err := odByteCount(res, 'j', "skip-bytes")
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "od: invalid skip argument: %s\n", err)
		return command.ExitUsage
	}
	length, hasLength, err := odOptionalCount(res, 'N', "read-bytes")
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "od: invalid length argument: %s\n", err)
		return command.ExitUsage
	}

	in, err := openDumpInput(ctx, "od", paths.ExpandGlobs(res.Positional))
	if err != nil {
		return command.ExitFailure
	}
	defer in.Close()

	if err := in.limit(skip, length, hasLength); err != nil {
		fmt.Fprintf(ctx.Stderr, "od: %s\n", err)
		return command.ExitFailure
	}

	if err := odDump(ctx, in.reader, format, radix, skip, res.Bool('v', "verbose")); err != nil {
		fmt.Fprintf(ctx.Stderr, "od: %s\n", err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func odDump(ctx *command.Context, r io.Reader, f odFormat, radix string, startOffset int64, verbose bool) error {
	lineBytes := f.size * f.perLine
	buf := make([]byte, lineBytes)
	// od counts offsets from the start of the file, so a -j skip is
	// included in the printed address rather than restarting at zero.
	offset := startOffset
	dw := newDumpWriter(ctx.Stdout, verbose)

	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			chunk := buf[:n]
			dw.row(chunk, odLine(chunk, f, radix, offset))
			offset += int64(n)
		}
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}
	}

	// The closing line is the total size, which is how od signals where
	// the dump ended. Suppressed when addresses are turned off.
	if radix != "n" {
		fmt.Fprintln(ctx.Stdout, odOffset(offset, radix))
	}
	return nil
}

func odLine(chunk []byte, f odFormat, radix string, offset int64) string {
	var sb strings.Builder
	if radix != "n" {
		sb.WriteString(odOffset(offset, radix))
	}
	for i := 0; i < len(chunk); i += f.size {
		end := i + f.size
		if end > len(chunk) {
			// A trailing fragment shorter than the item size is padded
			// with zero bytes, the same as od does.
			padded := make([]byte, f.size)
			copy(padded, chunk[i:])
			sb.WriteString(" ")
			sb.WriteString(f.render(padded))
			break
		}
		sb.WriteString(" ")
		sb.WriteString(f.render(chunk[i:end]))
	}
	return sb.String()
}

func odOffset(off int64, radix string) string {
	switch radix {
	case "d":
		return fmt.Sprintf("%07d", off)
	case "x":
		// GNU od pads hex addresses to six digits, not the seven it uses
		// for octal and decimal.
		return fmt.Sprintf("%06x", off)
	default:
		return fmt.Sprintf("%07o", off)
	}
}

// parseODFormat turns a -t argument such as "x2", "o", "c" or "u4" into
// a concrete renderer. An omitted size defaults to od's own per-type
// default (4 for integers, 1 for characters, 8 for floats).
func parseODFormat(spec string) (odFormat, error) {
	if spec == "" {
		return odFormat{}, fmt.Errorf("empty output format")
	}
	kind := spec[0]
	sizeText := spec[1:]

	size := 0
	if sizeText != "" {
		// Named sizes: C=char, S=short, I=int, L=long.
		switch strings.ToUpper(sizeText) {
		case "C":
			size = 1
		case "S":
			size = 2
		case "I":
			size = 4
		case "L":
			size = 8
		default:
			n, err := strconv.Atoi(sizeText)
			if err != nil {
				return odFormat{}, fmt.Errorf("invalid type string '%s'", spec)
			}
			size = n
		}
	}

	switch kind {
	case 'a':
		return odFormat{size: 1, perLine: 16, render: odNamedChar}, nil
	case 'c':
		return odFormat{size: 1, perLine: 16, render: odPrintableChar}, nil
	case 'o', 'x', 'd', 'u':
		if size == 0 {
			size = 4
		}
		if size != 1 && size != 2 && size != 4 && size != 8 {
			return odFormat{}, fmt.Errorf("invalid type string '%s'; size must be 1, 2, 4, or 8", spec)
		}
		return odIntegerFormat(kind, size), nil
	case 'f':
		if size == 0 {
			size = 8
		}
		switch size {
		case 4:
			return odFormat{size: 4, perLine: 4, render: func(b []byte) string {
				return fmt.Sprintf("%14.7e", math.Float32frombits(binary.LittleEndian.Uint32(b)))
			}}, nil
		case 8:
			return odFormat{size: 8, perLine: 2, render: func(b []byte) string {
				return fmt.Sprintf("%23.15e", math.Float64frombits(binary.LittleEndian.Uint64(b)))
			}}, nil
		default:
			return odFormat{}, fmt.Errorf("invalid type string '%s'; float size must be 4 or 8", spec)
		}
	default:
		return odFormat{}, fmt.Errorf("invalid type string '%s'", spec)
	}
}

// odIntegerFormat builds a renderer for the octal/hex/signed/unsigned
// integer views. Field widths match GNU od so columns line up with
// output people may already be diffing against.
func odIntegerFormat(kind byte, size int) odFormat {
	perLine := 16 / size
	if perLine < 1 {
		perLine = 1
	}
	read := func(b []byte) uint64 {
		switch size {
		case 1:
			return uint64(b[0])
		case 2:
			return uint64(binary.LittleEndian.Uint16(b))
		case 4:
			return uint64(binary.LittleEndian.Uint32(b))
		default:
			return binary.LittleEndian.Uint64(b)
		}
	}

	var width int
	switch kind {
	case 'o':
		width = map[int]int{1: 3, 2: 6, 4: 11, 8: 22}[size]
	case 'x':
		width = size * 2
	case 'd':
		width = map[int]int{1: 4, 2: 6, 4: 11, 8: 20}[size]
	case 'u':
		width = map[int]int{1: 3, 2: 5, 4: 10, 8: 20}[size]
	}

	return odFormat{size: size, perLine: perLine, render: func(b []byte) string {
		v := read(b)
		switch kind {
		case 'o':
			return fmt.Sprintf("%0*o", width, v)
		case 'x':
			return fmt.Sprintf("%0*x", width, v)
		case 'd':
			return fmt.Sprintf("%*d", width, signExtend(v, size))
		default:
			return fmt.Sprintf("%*d", width, v)
		}
	}}
}

// signExtend reinterprets a size-byte unsigned value as signed, which is
// what od -t d does.
func signExtend(v uint64, size int) int64 {
	switch size {
	case 1:
		return int64(int8(v))
	case 2:
		return int64(int16(v))
	case 4:
		return int64(int32(v))
	default:
		return int64(v)
	}
}

// odASCIINames are the ASCII control-character names od -t a prints.
var odASCIINames = [...]string{
	"nul", "soh", "stx", "etx", "eot", "enq", "ack", "bel",
	"bs", "ht", "nl", "vt", "ff", "cr", "so", "si",
	"dle", "dc1", "dc2", "dc3", "dc4", "nak", "syn", "etb",
	"can", "em", "sub", "esc", "fs", "gs", "rs", "us",
}

func odNamedChar(b []byte) string {
	c := b[0]
	switch {
	case c < 32:
		return fmt.Sprintf("%3s", odASCIINames[c])
	case c == 32:
		return " sp"
	case c == 127:
		return "del"
	case c < 127:
		return fmt.Sprintf("%3c", c)
	default:
		// Bytes above ASCII have no name; od prints them numerically.
		return fmt.Sprintf("%3o", c)
	}
}

// odEscapes are the C escape sequences od -t c prints in place of the
// control characters that have one.
var odEscapes = map[byte]string{
	0:    `\0`,
	'\a': `\a`,
	'\b': `\b`,
	'\f': `\f`,
	'\n': `\n`,
	'\r': `\r`,
	'\t': `\t`,
	'\v': `\v`,
}

func odPrintableChar(b []byte) string {
	c := b[0]
	if esc, ok := odEscapes[c]; ok {
		return fmt.Sprintf("%3s", esc)
	}
	if c >= 32 && c < 127 {
		return fmt.Sprintf("%3c", c)
	}
	return fmt.Sprintf("%3o", c)
}

// odByteCount reads a byte-count option, accepting the b/K/M/G suffixes
// od understands.
func odByteCount(res *parser.Result, short byte, long string) (int64, error) {
	v, ok := res.Value(short, long)
	if !ok {
		return 0, nil
	}
	return parseByteCount(v)
}

func odOptionalCount(res *parser.Result, short byte, long string) (int64, bool, error) {
	v, ok := res.Value(short, long)
	if !ok {
		return 0, false, nil
	}
	n, err := parseByteCount(v)
	return n, true, err
}

// parseByteCount accepts plain decimal, 0x-prefixed hex, 0-prefixed
// octal, and the b/k/m/g multipliers od and hexdump both allow.
func parseByteCount(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty count")
	}
	mult := int64(1)
	switch s[len(s)-1] {
	case 'b', 'B':
		mult, s = 512, s[:len(s)-1]
	case 'k', 'K':
		mult, s = 1024, s[:len(s)-1]
	case 'm', 'M':
		mult, s = 1024*1024, s[:len(s)-1]
	case 'g', 'G':
		mult, s = 1024*1024*1024, s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number '%s'", s)
	}
	if n < 0 {
		return 0, fmt.Errorf("count cannot be negative")
	}
	return n * mult, nil
}

func init() { command.Register(odCommand{}) }
