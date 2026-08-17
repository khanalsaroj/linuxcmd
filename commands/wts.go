package commands

import (
	"syscall"
	"unsafe"
)

// wtsSessionInfo mirrors the Win32 WTS_SESSION_INFOW struct returned by
// WTSEnumerateSessionsW.
type wtsSessionInfo struct {
	SessionID      uint32
	WinStationName *uint16
	State          uint32
}

const wtsUserNameInfoClass = 5 // WTSUserName

// enumerateSessionUsers lists the usernames of interactive Windows
// sessions via the Terminal Services API, the native way to enumerate
// logged-in users (what "who"/"users"/"w" are built on). It can return an
// empty list (not an error) on machines with no interactive sessions, and
// an error if the API itself is unavailable or access is denied.
//
// Pointers returned by the API are kept as unsafe.Pointer (never round
// tripped through uintptr) and walked with unsafe.Add, the vet-sanctioned
// way to do pointer arithmetic on foreign, non-Go-GC-managed memory.
func enumerateSessionUsers() ([]string, error) {
	wtsapi32 := syscall.NewLazyDLL("wtsapi32.dll")
	enumProc := wtsapi32.NewProc("WTSEnumerateSessionsW")
	queryProc := wtsapi32.NewProc("WTSQuerySessionInformationW")
	freeProc := wtsapi32.NewProc("WTSFreeMemory")

	var sessionInfo unsafe.Pointer
	var count uint32
	ret, _, err := enumProc.Call(0, 0, 1, uintptr(unsafe.Pointer(&sessionInfo)), uintptr(unsafe.Pointer(&count)))
	if ret == 0 {
		return nil, err
	}
	defer freeProc.Call(uintptr(sessionInfo))

	var names []string
	entrySize := unsafe.Sizeof(wtsSessionInfo{})
	for i := uint32(0); i < count; i++ {
		entry := (*wtsSessionInfo)(unsafe.Add(sessionInfo, uintptr(i)*entrySize))

		var buf unsafe.Pointer
		var bytesReturned uint32
		qret, _, _ := queryProc.Call(0, uintptr(entry.SessionID), wtsUserNameInfoClass,
			uintptr(unsafe.Pointer(&buf)), uintptr(unsafe.Pointer(&bytesReturned)))
		if qret == 0 || buf == nil {
			continue
		}
		length := bytesReturned / 2
		if length > 0 {
			slice := unsafe.Slice((*uint16)(buf), length)
			name := syscall.UTF16ToString(slice)
			if name != "" {
				names = append(names, name)
			}
		}
		freeProc.Call(uintptr(buf))
	}
	return names, nil
}
