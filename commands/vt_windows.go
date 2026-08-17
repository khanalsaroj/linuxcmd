package commands

import (
	"syscall"
	"unsafe"
)

const enableVirtualTerminalProcessing = 0x0004

// enableVirtualTerminal turns on ANSI escape sequence interpretation for
// the current stdout console handle. Best-effort: if stdout isn't a
// console (e.g. redirected to a file) or the OS call fails, it silently
// does nothing, since ANSI codes written to a file are harmless anyway.
func enableVirtualTerminal() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	const stdOutputHandle = ^uint32(10) // -11 as uint32 (STD_OUTPUT_HANDLE)
	h, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	if h == 0 || h == ^uintptr(0) {
		return
	}

	var mode uint32
	ret, _, _ := getConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return
	}
	_, _, _ = setConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
}
