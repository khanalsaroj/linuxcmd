package commands

import (
	"strings"
	"testing"
)

func TestOdDefaultIsOctalWords(t *testing.T) {
	code, out, errOut := runWithStdin(t, "od", "AB")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	// "AB" little-endian as one 16-bit word is 0x4241 == 041101 octal.
	if !strings.Contains(out, "041101") {
		t.Errorf("od output = %q, want the octal word 041101", out)
	}
	if !strings.HasPrefix(out, "0000000") {
		t.Errorf("od output = %q, want a 7-digit octal offset first", out)
	}
}

func TestOdCharacterFormat(t *testing.T) {
	code, out, errOut := runWithStdin(t, "od", "A\nB", "-c")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, `\n`) {
		t.Errorf("od -c output = %q, want the newline shown as an escape", out)
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("od -c output = %q, want the printable characters", out)
	}
}

func TestOdHexAddressAndType(t *testing.T) {
	code, out, errOut := runWithStdin(t, "od", "hello", "-A", "x", "-t", "x1")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	// GNU od pads hex addresses to six digits, not seven.
	if !strings.HasPrefix(out, "000000 ") {
		t.Errorf("od -A x output = %q, want a 6-digit hex offset", out)
	}
	if !strings.Contains(out, "68 65 6c 6c 6f") {
		t.Errorf("od -t x1 output = %q, want the bytes of 'hello'", out)
	}
}

func TestOdSkipAndLimit(t *testing.T) {
	code, out, errOut := runWithStdin(t, "od", "0123456789", "-A", "d", "-t", "c", "-j", "4", "-N", "2")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "4") || !strings.Contains(out, "5") {
		t.Errorf("od -j 4 -N 2 output = %q, want only the bytes '4' and '5'", out)
	}
	if strings.Contains(out, "  3 ") || strings.Contains(out, "  6 ") {
		t.Errorf("od -j 4 -N 2 output = %q, leaked bytes outside the window", out)
	}
	// Offsets continue from the start of input, so the first line is at 4.
	if !strings.HasPrefix(out, "0000004") {
		t.Errorf("od output = %q, want the address to include the skipped bytes", out)
	}
}

// Runs of identical lines collapse to "*" unless -v is given.
func TestOdCollapsesRepeatedLines(t *testing.T) {
	input := strings.Repeat("A", 64)

	_, collapsed, _ := runWithStdin(t, "od", input, "-t", "x1")
	if !strings.Contains(collapsed, "*") {
		t.Errorf("od output = %q, want repeated lines collapsed to '*'", collapsed)
	}

	_, verbose, _ := runWithStdin(t, "od", input, "-t", "x1", "-v")
	if strings.Contains(verbose, "*") {
		t.Errorf("od -v output = %q, want no collapsing", verbose)
	}
	if strings.Count(verbose, "\n") <= strings.Count(collapsed, "\n") {
		t.Error("od -v should produce more lines than the collapsed form")
	}
}

func TestOdRejectsBadType(t *testing.T) {
	code, _, errOut := runWithStdin(t, "od", "x", "-t", "q9")
	if code == 0 {
		t.Error("expected a nonzero exit for an invalid type string")
	}
	if !strings.Contains(errOut, "invalid type string") {
		t.Errorf("stderr = %q, want an invalid-type message", errOut)
	}
}

func TestHexdumpCanonical(t *testing.T) {
	code, out, errOut := runWithStdin(t, "hexdump", "hello world", "-C")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "68 65 6c 6c 6f") {
		t.Errorf("hexdump -C output = %q, want hex bytes", out)
	}
	if !strings.Contains(out, "|hello world|") {
		t.Errorf("hexdump -C output = %q, want the printable gutter", out)
	}
	if !strings.HasPrefix(out, "00000000") {
		t.Errorf("hexdump -C output = %q, want an 8-digit offset", out)
	}
}

func TestHexdumpDefaultIsTwoByteHex(t *testing.T) {
	code, out, errOut := runWithStdin(t, "hexdump", "AB")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	// Little-endian 16-bit word for "AB".
	if !strings.Contains(out, "4241") {
		t.Errorf("hexdump output = %q, want the little-endian word 4241", out)
	}
}

func TestHexdumpLengthAndSkip(t *testing.T) {
	code, out, errOut := runWithStdin(t, "hexdump", "abcdefgh", "-C", "-s", "2", "-n", "3")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, "|cde|") {
		t.Errorf("hexdump -s 2 -n 3 output = %q, want only 'cde'", out)
	}
}

func TestHexdumpCharacterFormat(t *testing.T) {
	code, out, errOut := runWithStdin(t, "hexdump", "a\tb", "-c")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(out, `\t`) {
		t.Errorf("hexdump -c output = %q, want the tab shown as an escape", out)
	}
}

func TestDumpCommandsReportMissingFile(t *testing.T) {
	for _, name := range []string{"od", "hexdump"} {
		code, _, errOut := run(t, name, "no-such-file-here.bin")
		if code == 0 {
			t.Errorf("%s: expected a nonzero exit for a missing file", name)
		}
		if !strings.Contains(errOut, "No such file or directory") {
			t.Errorf("%s: stderr = %q, want a Linux-style not-found message", name, errOut)
		}
	}
}
