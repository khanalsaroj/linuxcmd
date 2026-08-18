package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// xclip and xsel bridge shell pipelines to the Windows clipboard. On
// Linux these talk to an X11 selection owner; here they talk to the Win32
// clipboard, which is the direct equivalent and needs no X server.
//
// One behavioral difference is worth knowing: X11 has three independent
// selections (PRIMARY, SECONDARY, CLIPBOARD) and Windows has exactly one
// clipboard, so -selection and xsel's -p/-s/-b all address the same
// storage. Scripts that use PRIMARY and CLIPBOARD to hold two different
// values will see them collapse into one.

// Win32 clipboard constants.
const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	kernel32Clip          = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard     = user32.NewProc("OpenClipboard")
	procCloseClipboard    = user32.NewProc("CloseClipboard")
	procEmptyClipboard    = user32.NewProc("EmptyClipboard")
	procGetClipboardData  = user32.NewProc("GetClipboardData")
	procSetClipboardData  = user32.NewProc("SetClipboardData")
	procIsClipboardFormat = user32.NewProc("IsClipboardFormatAvailable")
	procGlobalAlloc       = kernel32Clip.NewProc("GlobalAlloc")
	procGlobalFree        = kernel32Clip.NewProc("GlobalFree")
	procGlobalLock        = kernel32Clip.NewProc("GlobalLock")
	procGlobalUnlock      = kernel32Clip.NewProc("GlobalUnlock")
	procGlobalSize        = kernel32Clip.NewProc("GlobalSize")
	procMoveMemory        = kernel32Clip.NewProc("RtlMoveMemory")
)

// copyFromGlobal and copyToGlobal move bytes between a Go slice and the
// global memory block the clipboard uses. Both sides are passed to
// RtlMoveMemory as addresses, so the block's address stays a uintptr and
// is never converted into a Go pointer -- doing that conversion would be
// unsound (the address is not Go-managed memory) and go vet rejects it.

func copyFromGlobal(dst []uint16, src uintptr, byteCount uintptr) {
	if len(dst) == 0 || byteCount == 0 {
		return
	}
	procMoveMemory.Call(uintptr(unsafe.Pointer(&dst[0])), src, byteCount)
}

func copyToGlobal(dst uintptr, src []uint16) {
	if len(src) == 0 {
		return
	}
	procMoveMemory.Call(dst, uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)*2))
}

// openClipboard takes ownership of the clipboard, retrying briefly. Any
// other process can hold it open momentarily (Explorer and Office do
// this routinely), and a single failed attempt is not a real error.
func openClipboard() error {
	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		ret, _, err := procOpenClipboard.Call(0)
		if ret != 0 {
			return nil
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("clipboard is in use by another program: %w", lastErr)
}

// readClipboard returns the clipboard's Unicode text, or "" when it
// holds no text at all (an image, say, or nothing).
func readClipboard() (string, error) {
	if err := openClipboard(); err != nil {
		return "", err
	}
	defer procCloseClipboard.Call()

	available, _, _ := procIsClipboardFormat.Call(cfUnicodeText)
	if available == 0 {
		return "", nil
	}

	h, _, err := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", fmt.Errorf("cannot read clipboard: %w", err)
	}
	size, _, _ := procGlobalSize.Call(h)
	if size == 0 {
		return "", nil
	}

	p, _, err := procGlobalLock.Call(h)
	if p == 0 {
		return "", fmt.Errorf("cannot lock clipboard memory: %w", err)
	}
	defer procGlobalUnlock.Call(h)

	units := make([]uint16, size/2)
	copyFromGlobal(units, p, size)
	// UTF16ToString stops at the first NUL, which is where the clipboard
	// text ends; GlobalSize reports the whole allocation, which may be
	// larger.
	return syscall.UTF16ToString(units), nil
}

// writeClipboard replaces the clipboard's contents with text.
func writeClipboard(text string) error {
	units, err := syscall.UTF16FromString(text)
	if err != nil {
		return fmt.Errorf("text contains a NUL byte and cannot be placed on the clipboard")
	}

	if err := openClipboard(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	if ret, _, err := procEmptyClipboard.Call(); ret == 0 {
		return fmt.Errorf("cannot clear clipboard: %w", err)
	}

	size := uintptr(len(units) * 2)
	h, _, err := procGlobalAlloc.Call(gmemMoveable, size)
	if h == 0 {
		return fmt.Errorf("cannot allocate clipboard memory: %w", err)
	}

	p, _, err := procGlobalLock.Call(h)
	if p == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("cannot lock clipboard memory: %w", err)
	}
	copyToGlobal(p, units)
	procGlobalUnlock.Call(h)

	// Ownership of the memory transfers to the system on success; on
	// failure it is still ours to release.
	if ret, _, err := procSetClipboardData.Call(cfUnicodeText, h); ret == 0 {
		procGlobalFree.Call(h)
		return fmt.Errorf("cannot write clipboard: %w", err)
	}
	return nil
}

// --- xclip ---------------------------------------------------------------

type xclipCommand struct{}

func (xclipCommand) Name() string    { return "xclip" }
func (xclipCommand) Summary() string { return "copy to or paste from the Windows clipboard" }

var xclipSpec = parser.Spec{
	{Short: 'i', Long: "in"},
	{Short: 'o', Long: "out"},
	{Short: 'r', Long: "rmlastnl"},
	{Long: "selection", HasArg: true},
	{Long: "display", HasArg: true}, // accepted and ignored; no X server here
}

func (xclipCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, xclipSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "xclip: %s\n", err)
		return command.ExitUsage
	}

	if res.Bool('o', "out") {
		return clipboardOut(ctx, "xclip")
	}
	return clipboardIn(ctx, "xclip", res.Positional, res.Bool('r', "rmlastnl"))
}

// --- xsel ----------------------------------------------------------------

type xselCommand struct{}

func (xselCommand) Name() string    { return "xsel" }
func (xselCommand) Summary() string { return "read or set the Windows clipboard" }

var xselSpec = parser.Spec{
	{Short: 'i', Long: "input"},
	{Short: 'o', Long: "output"},
	{Short: 'c', Long: "clear"},
	{Short: 'b', Long: "clipboard"},
	{Short: 'p', Long: "primary"},
	{Short: 's', Long: "secondary"},
	{Short: 'n', Long: "nodetach"},
}

func (xselCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, xselSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "xsel: %s\n", err)
		return command.ExitUsage
	}

	if res.Bool('c', "clear") {
		if err := writeClipboard(""); err != nil {
			fmt.Fprintf(ctx.Stderr, "xsel: %s\n", err)
			return command.ExitFailure
		}
		return command.ExitSuccess
	}
	if res.Bool('i', "input") {
		return clipboardIn(ctx, "xsel", res.Positional, false)
	}
	// Unlike xclip, xsel with no direction flag prints the selection.
	return clipboardOut(ctx, "xsel")
}

// --- shared --------------------------------------------------------------

func clipboardOut(ctx *command.Context, prog string) int {
	text, err := readClipboard()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}
	if _, err := io.WriteString(ctx.Stdout, text); err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// clipboardIn loads the clipboard from the named files, or from standard
// input when no file is given, which is the piped form these tools are
// almost always used in.
func clipboardIn(ctx *command.Context, prog string, operands []string, trimTrailingNewline bool) int {
	var sb strings.Builder

	files := paths.ExpandGlobs(operands)
	if len(files) == 0 {
		data, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
			return command.ExitFailure
		}
		sb.Write(data)
	} else {
		for _, name := range files {
			resolved, err := paths.Resolve(name)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, prog, name, err)
				return command.ExitFailure
			}
			data, err := os.ReadFile(resolved)
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, prog, name, err)
				return command.ExitFailure
			}
			sb.Write(data)
		}
	}

	text := sb.String()
	if trimTrailingNewline {
		text = strings.TrimRight(text, "\r\n")
	}
	if err := writeClipboard(text); err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", prog, err)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() {
	command.Register(xclipCommand{})
	command.Register(xselCommand{})
}
