package commands

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

// Shared Windows volume enumeration for lsblk and blkid. Linux exposes
// block devices through /sys and /dev; Windows exposes them through a
// handful of kernel32 calls plus one storage ioctl, and this file is the
// translation layer between the two views.

const (
	driveUnknown   = 0
	driveNoRootDir = 1
	driveRemovable = 2
	driveFixed     = 3
	driveRemote    = 4
	driveCDROM     = 5
	driveRAMDisk   = 6
)

// ioctlStorageGetDeviceNumber maps a volume back to the physical disk it
// lives on, which is what lets lsblk group partitions under a disk.
const ioctlStorageGetDeviceNumber = 0x2D1080

// ioctlDiskGetLengthInfo reports a physical disk's true capacity,
// including space not covered by any volume.
const ioctlDiskGetLengthInfo = 0x7405C

var (
	kernel32Vol             = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDriveStrs = kernel32Vol.NewProc("GetLogicalDriveStringsW")
	procGetDriveType        = kernel32Vol.NewProc("GetDriveTypeW")
	procGetVolumeInfo       = kernel32Vol.NewProc("GetVolumeInformationW")
	procGetDiskFreeSpaceEx  = kernel32Vol.NewProc("GetDiskFreeSpaceExW")
)

// volumeInfo is one mounted Windows volume, described in the terms
// lsblk and blkid need.
type volumeInfo struct {
	Letter    string // "C:"
	Root      string // "C:\"
	Label     string
	FSType    string // NTFS, FAT32, exFAT, ReFS
	Serial    uint32
	Total     uint64
	Free      uint64
	DriveType uint32
	// DiskNumber is the physical disk this volume sits on, or -1 when it
	// cannot be determined (network shares and some virtual volumes).
	DiskNumber      int
	PartitionNumber int
}

// Removable reports whether the volume is on removable media, which is
// lsblk's RM column.
func (v volumeInfo) Removable() bool {
	return v.DriveType == driveRemovable || v.DriveType == driveCDROM
}

// TypeName maps the Windows drive type onto lsblk's TYPE vocabulary.
func (v volumeInfo) TypeName() string {
	switch v.DriveType {
	case driveCDROM:
		return "rom"
	case driveRemote:
		return "network"
	case driveRAMDisk:
		return "ram"
	default:
		return "part"
	}
}

// enumerateVolumes lists every volume that currently has a drive letter.
// Volumes without one (recovery and EFI system partitions) are not
// reachable through GetLogicalDriveStrings and are therefore absent;
// that limitation is documented rather than worked around, because
// enumerating them requires walking volume GUID paths that no other
// command in this project needs.
func enumerateVolumes() ([]volumeInfo, error) {
	buf := make([]uint16, 256)
	ret, _, err := procGetLogicalDriveStrs.Call(uintptr(len(buf)), uintptr(unsafe.Pointer(&buf[0])))
	if ret == 0 {
		return nil, fmt.Errorf("cannot list drives: %w", err)
	}

	var vols []volumeInfo
	for _, root := range splitNulSeparated(buf[:ret]) {
		if root == "" {
			continue
		}
		v := volumeInfo{
			Root:            root,
			Letter:          strings.TrimSuffix(root, `\`),
			DiskNumber:      -1,
			PartitionNumber: -1,
		}

		rootPtr, err := syscall.UTF16PtrFromString(root)
		if err != nil {
			continue
		}
		dt, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(rootPtr)))
		v.DriveType = uint32(dt)
		if v.DriveType == driveNoRootDir || v.DriveType == driveUnknown {
			continue
		}

		v.Label, v.FSType, v.Serial = volumeIdentity(rootPtr)
		v.Total, v.Free = volumeCapacity(rootPtr)
		v.DiskNumber, v.PartitionNumber = volumeDiskLocation(v.Letter)

		vols = append(vols, v)
	}
	return vols, nil
}

// volumeIdentity reads the label, filesystem name and serial number. An
// empty CD-ROM drive fails here, which is expected rather than an error.
func volumeIdentity(rootPtr *uint16) (label, fsType string, serial uint32) {
	nameBuf := make([]uint16, 261)
	fsBuf := make([]uint16, 261)
	var maxComponent, flags uint32

	ret, _, _ := procGetVolumeInfo.Call(
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&nameBuf[0])), uintptr(len(nameBuf)),
		uintptr(unsafe.Pointer(&serial)),
		uintptr(unsafe.Pointer(&maxComponent)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&fsBuf[0])), uintptr(len(fsBuf)),
	)
	if ret == 0 {
		return "", "", 0
	}
	return syscall.UTF16ToString(nameBuf), syscall.UTF16ToString(fsBuf), serial
}

func volumeCapacity(rootPtr *uint16) (total, free uint64) {
	var availableToCaller uint64
	ret, _, _ := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&availableToCaller)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&free)),
	)
	if ret == 0 {
		return 0, 0
	}
	return total, free
}

// storageDeviceNumber mirrors the Win32 STORAGE_DEVICE_NUMBER struct.
type storageDeviceNumber struct {
	DeviceType      uint32
	DeviceNumber    uint32
	PartitionNumber uint32
}

// volumeDiskLocation asks the storage stack which physical disk and
// partition back a drive letter. The volume is opened with no access
// rights at all, which is enough for this query and avoids needing
// administrator privileges.
func volumeDiskLocation(letter string) (disk, partition int) {
	path, err := syscall.UTF16PtrFromString(`\\.\` + letter)
	if err != nil {
		return -1, -1
	}
	handle, err := syscall.CreateFile(
		path,
		0, // query only
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return -1, -1
	}
	defer syscall.CloseHandle(handle)

	var out storageDeviceNumber
	var returned uint32
	err = syscall.DeviceIoControl(
		handle,
		ioctlStorageGetDeviceNumber,
		nil, 0,
		(*byte)(unsafe.Pointer(&out)), uint32(unsafe.Sizeof(out)),
		&returned, nil,
	)
	if err != nil {
		return -1, -1
	}
	return int(out.DeviceNumber), int(out.PartitionNumber)
}

// physicalDiskSize reports a disk's capacity in bytes, or 0 when it
// cannot be read. Callers fall back to summing the disk's volumes.
func physicalDiskSize(diskNumber int) uint64 {
	path, err := syscall.UTF16PtrFromString(fmt.Sprintf(`\\.\PhysicalDrive%d`, diskNumber))
	if err != nil {
		return 0
	}
	handle, err := syscall.CreateFile(
		path,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return 0
	}
	defer syscall.CloseHandle(handle)

	var length uint64
	var returned uint32
	if err := syscall.DeviceIoControl(
		handle,
		ioctlDiskGetLengthInfo,
		nil, 0,
		(*byte)(unsafe.Pointer(&length)), uint32(unsafe.Sizeof(length)),
		&returned, nil,
	); err != nil {
		return 0
	}
	return length
}

// splitNulSeparated decodes the NUL-separated, double-NUL-terminated
// string list several kernel32 enumeration APIs return.
func splitNulSeparated(buf []uint16) []string {
	var out []string
	start := 0
	for i, u := range buf {
		if u != 0 {
			continue
		}
		if i > start {
			out = append(out, syscall.UTF16ToString(buf[start:i]))
		}
		start = i + 1
	}
	return out
}
