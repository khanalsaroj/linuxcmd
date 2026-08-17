package commands

import (
	"bytes"
	"fmt"
	"os"
	"unicode/utf8"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type fileCommand struct{}

func (fileCommand) Name() string    { return "file" }
func (fileCommand) Summary() string { return "identify a file's type" }

func (fileCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: file FILE...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(ctx.Args) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "file", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "file", arg, err)
			exit = command.ExitFailure
			continue
		}
		if info.IsDir() {
			fmt.Fprintf(ctx.Stdout, "%s: directory\n", arg)
			continue
		}

		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "file", arg, err)
			exit = command.ExitFailure
			continue
		}
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		f.Close()
		fmt.Fprintf(ctx.Stdout, "%s: %s\n", arg, describeFile(buf[:n], info.Size()))
	}
	return exit
}

// describeFile applies a small set of magic-byte signatures, falling back
// to a text/binary heuristic based on printable-character ratio.
func describeFile(head []byte, size int64) string {
	if size == 0 {
		return "empty"
	}
	switch {
	case bytes.HasPrefix(head, []byte("\x89PNG\r\n\x1a\n")):
		return "PNG image data"
	case bytes.HasPrefix(head, []byte("\xff\xd8\xff")):
		return "JPEG image data"
	case bytes.HasPrefix(head, []byte("GIF87a")), bytes.HasPrefix(head, []byte("GIF89a")):
		return "GIF image data"
	case bytes.HasPrefix(head, []byte("PK\x03\x04")):
		return "Zip archive data"
	case bytes.HasPrefix(head, []byte("\x1f\x8b")):
		return "gzip compressed data"
	case bytes.HasPrefix(head, []byte("%PDF-")):
		return "PDF document"
	case bytes.HasPrefix(head, []byte("MZ")):
		return "PE32 executable (Windows)"
	case bytes.HasPrefix(head, []byte("\x7fELF")):
		return "ELF executable"
	case bytes.HasPrefix(head, []byte{0xca, 0xfe, 0xba, 0xbe}):
		return "Mach-O/Java class data"
	}
	if isText(head) {
		return "ASCII text"
	}
	return "data"
}

func isText(head []byte) bool {
	if len(head) == 0 {
		return true
	}
	printable := 0
	for i := 0; i < len(head); {
		r, size := utf8.DecodeRune(head[i:])
		if r == utf8.RuneError && size == 1 {
			return false
		}
		if r == 0 {
			return false
		}
		if r == '\n' || r == '\r' || r == '\t' || (r >= 0x20 && r < 0x7f) || r > 0x7f {
			printable++
		}
		i += size
	}
	return printable*10 >= len(head)*9
}

func init() { command.Register(fileCommand{}) }
