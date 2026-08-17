package commands

import (
	"syscall"
	"unsafe"
)

// procInfo is one row of a process snapshot, shared by ps, pgrep, pkill,
// and pstree so they don't each re-implement the Tool Help API dance.
type procInfo struct {
	PID, PPID uint32
	Name      string
}

func snapshotProcesses() ([]procInfo, error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(th32csSnapProcess, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(snapshot)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	var procs []procInfo
	for err = syscall.Process32First(snapshot, &entry); err == nil; err = syscall.Process32Next(snapshot, &entry) {
		procs = append(procs, procInfo{
			PID:  entry.ProcessID,
			PPID: entry.ParentProcessID,
			Name: syscall.UTF16ToString(entry.ExeFile[:]),
		})
	}
	return procs, nil
}
