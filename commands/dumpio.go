package commands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// Shared input plumbing for the byte-dumping commands (od, hexdump).
// Both concatenate their operands into a single stream, both support
// skipping a prefix and limiting the length, and both collapse runs of
// identical output lines to a single "*" unless asked not to. Keeping
// that here means od and hexdump can differ only in how they format a
// row of bytes, which is the only place they actually disagree.

// dumpInput is a concatenated view of every operand, with the file
// handles it owns so the caller can release them.
type dumpInput struct {
	reader io.Reader
	closers []io.Closer
}

// Close releases every file opened for this input.
func (d *dumpInput) Close() {
	for _, c := range d.closers {
		_ = c.Close()
	}
}

// openDumpInput resolves each operand and concatenates them, matching
// the way od and hexdump treat multiple files as one continuous byte
// stream. With no operands (or a lone "-") it reads standard input.
func openDumpInput(ctx *command.Context, prog string, operands []string) (*dumpInput, error) {
	in := &dumpInput{}
	if len(operands) == 0 {
		in.reader = ctx.Stdin
		return in, nil
	}

	var readers []io.Reader
	for _, name := range operands {
		if name == "-" {
			readers = append(readers, ctx.Stdin)
			continue
		}
		resolved, err := paths.Resolve(name)
		if err != nil {
			in.Close()
			output.SimpleErrorf(ctx.Stderr, prog, name, err)
			return nil, err
		}
		f, err := os.Open(resolved)
		if err != nil {
			in.Close()
			output.SimpleErrorf(ctx.Stderr, prog, name, err)
			return nil, err
		}
		in.closers = append(in.closers, f)
		readers = append(readers, f)
	}
	in.reader = io.MultiReader(readers...)
	return in, nil
}

// limit applies the -j/-s skip and -N/-n length restrictions. Skipping
// is done by reading and discarding rather than seeking, because the
// stream may be stdin or a concatenation of several files, neither of
// which is seekable as a whole.
func (d *dumpInput) limit(skip, length int64, hasLength bool) error {
	if skip > 0 {
		if _, err := io.CopyN(io.Discard, d.reader, skip); err != nil {
			if errors.Is(err, io.EOF) {
				// Skipping past the end yields an empty dump, which is
				// what both tools do; not an error.
				d.reader = eofReader{}
				return nil
			}
			return err
		}
	}
	if hasLength {
		d.reader = io.LimitReader(d.reader, length)
	}
	return nil
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) { return 0, io.EOF }

// dumpWriter emits formatted rows while collapsing consecutive identical
// rows into a single "*", the behavior both od and hexdump use to keep
// output short over long runs of repeated bytes. Setting verbose
// disables the collapsing (od -v, hexdump -v).
type dumpWriter struct {
	w         io.Writer
	verbose   bool
	lastBytes string
	repeating bool
	haveLast  bool
}

func newDumpWriter(w io.Writer, verbose bool) *dumpWriter {
	return &dumpWriter{w: w, verbose: verbose}
}

// row writes one formatted line for the given chunk of source bytes.
// The chunk (not the rendered text) is the identity used for collapsing,
// matching od/hexdump, which compare input rather than output.
func (d *dumpWriter) row(chunk []byte, line string) {
	if !d.verbose {
		key := string(chunk)
		if d.haveLast && key == d.lastBytes {
			if !d.repeating {
				fmt.Fprintln(d.w, "*")
				d.repeating = true
			}
			return
		}
		d.lastBytes = key
		d.haveLast = true
		d.repeating = false
	}
	fmt.Fprintln(d.w, line)
}
