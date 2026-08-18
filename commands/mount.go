package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// mount and umount cover the part of the Linux idea that Windows
// actually has. Windows has no single directory tree to graft a
// filesystem into: local volumes are assigned drive letters by the
// storage stack, and network shares are attached to drive letters by the
// redirector. So "list what is mounted" translates exactly, and
// "attach //server/share to Z:" translates exactly, while "mount an
// image at /mnt/foo" has no equivalent and is reported as unsupported
// rather than approximated.
type mountCommand struct{}

func (mountCommand) Name() string    { return "mount" }
func (mountCommand) Summary() string { return "list mounted volumes or attach a network share" }

var mountSpec = parser.Spec{
	{Short: 't', HasArg: true},
	{Short: 'o', HasArg: true},
	{Short: 'a', Long: "all"},
	{Short: 'v', Long: "verbose"},
}

var (
	mpr                    = syscall.NewLazyDLL("mpr.dll")
	procWNetGetConnection  = mpr.NewProc("WNetGetConnectionW")
	procWNetAddConnection2 = mpr.NewProc("WNetAddConnection2W")
	procWNetCancelConn2    = mpr.NewProc("WNetCancelConnection2W")
)

func (mountCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, mountSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mount: %s\n", err)
		return command.ExitUsage
	}

	switch len(res.Positional) {
	case 0:
		return listMounts(ctx)
	case 2:
		return attachShare(ctx, res.Positional[0], res.Positional[1])
	default:
		fmt.Fprintln(ctx.Stderr, "usage: mount                      list mounted volumes")
		fmt.Fprintln(ctx.Stderr, "       mount //server/share DRIVE attach a network share")
		return command.ExitUsage
	}
}

func listMounts(ctx *command.Context) int {
	vols, err := enumerateVolumes()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mount: %s\n", err)
		return command.ExitFailure
	}

	for _, v := range vols {
		source := v.Letter
		fsType := v.FSType
		if fsType == "" {
			fsType = "none"
		}

		// A mapped network drive's real source is the UNC path behind
		// it, which is the useful thing to show.
		if v.DriveType == driveRemote {
			if unc, ok := networkPathFor(v.Letter); ok {
				source = unc
			}
			// Linux reports SMB mounts as cifs, which is the closest
			// name for what Windows attaches here.
			fsType = "cifs"
		}

		options := "rw"
		if v.DriveType == driveCDROM {
			options = "ro"
		}
		if v.Removable() {
			options += ",removable"
		}
		fmt.Fprintf(ctx.Stdout, "%s on %s type %s (%s)\n", source, v.Root, fsType, options)
	}
	return command.ExitSuccess
}

// networkPathFor returns the UNC path a mapped drive letter points at.
func networkPathFor(letter string) (string, bool) {
	local, err := syscall.UTF16PtrFromString(letter)
	if err != nil {
		return "", false
	}
	buf := make([]uint16, 1024)
	length := uint32(len(buf))
	ret, _, _ := procWNetGetConnection.Call(
		uintptr(unsafe.Pointer(local)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
	)
	if ret != 0 {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}

// netResource mirrors the Win32 NETRESOURCE struct used to describe a
// share to WNetAddConnection2.
type netResource struct {
	Scope       uint32
	Type        uint32
	DisplayType uint32
	Usage       uint32
	LocalName   *uint16
	RemoteName  *uint16
	Comment     *uint16
	Provider    *uint16
}

const resourceTypeDisk = 0x00000001

func attachShare(ctx *command.Context, source, target string) int {
	unc, ok := uncPath(source)
	if !ok {
		fmt.Fprintf(ctx.Stderr, "mount: %s: only network shares can be mounted\n", source)
		fmt.Fprintln(ctx.Stderr, "mount: give a share as //server/share or \\\\server\\share")
		fmt.Fprintln(ctx.Stderr, "mount: to mount a disk image use PowerShell's Mount-DiskImage")
		return command.ExitFailure
	}

	drive := strings.TrimSuffix(strings.ToUpper(target), `\`)
	if len(drive) != 2 || drive[1] != ':' {
		fmt.Fprintf(ctx.Stderr, "mount: %s: a network share must be attached to a drive letter such as Z:\n", target)
		return command.ExitFailure
	}

	localPtr, err := syscall.UTF16PtrFromString(drive)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mount: %s\n", err)
		return command.ExitFailure
	}
	remotePtr, err := syscall.UTF16PtrFromString(unc)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "mount: %s\n", err)
		return command.ExitFailure
	}

	resource := netResource{
		Type:       resourceTypeDisk,
		LocalName:  localPtr,
		RemoteName: remotePtr,
	}
	// Credentials are left nil so Windows uses the current user's
	// context or prompts through its own UI; linuxcmd never handles
	// passwords itself.
	ret, _, _ := procWNetAddConnection2.Call(
		uintptr(unsafe.Pointer(&resource)),
		0, // no password
		0, // no username
		0, // not persistent across logons
	)
	if ret != 0 {
		fmt.Fprintf(ctx.Stderr, "mount: cannot attach %s to %s: %s\n",
			unc, drive, syscall.Errno(ret))
		return command.ExitFailure
	}
	return command.ExitSuccess
}

// uncPath normalizes a share given in either Linux (//server/share) or
// Windows (\\server\share) spelling.
func uncPath(s string) (string, bool) {
	switch {
	case strings.HasPrefix(s, "//"):
		return `\\` + filepath.FromSlash(strings.TrimPrefix(s, "//")), true
	case strings.HasPrefix(s, `\\`):
		return s, true
	default:
		return "", false
	}
}

// --- umount --------------------------------------------------------------

type umountCommand struct{}

func (umountCommand) Name() string    { return "umount" }
func (umountCommand) Summary() string { return "detach a mapped network drive" }

var umountSpec = parser.Spec{
	{Short: 'f', Long: "force"},
	{Short: 'v', Long: "verbose"},
}

func (umountCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, umountSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "umount: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: umount DRIVE")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, target := range res.Positional {
		drive := strings.TrimSuffix(strings.ToUpper(target), `\`)
		if len(drive) != 2 || drive[1] != ':' {
			fmt.Fprintf(ctx.Stderr, "umount: %s: not a drive letter\n", target)
			exit = command.ExitFailure
			continue
		}

		localPtr, err := syscall.UTF16PtrFromString(drive)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "umount: %s\n", err)
			exit = command.ExitFailure
			continue
		}
		flags := uintptr(0)
		force := uintptr(0)
		if res.Bool('f', "force") {
			// Disconnect even with open files, which is the closest
			// equivalent to umount -f.
			force = 1
		}
		ret, _, _ := procWNetCancelConn2.Call(uintptr(unsafe.Pointer(localPtr)), flags, force)
		if ret != 0 {
			fmt.Fprintf(ctx.Stderr, "umount: cannot detach %s: %s\n", drive, syscall.Errno(ret))
			fmt.Fprintln(ctx.Stderr, "umount: local volumes are detached with mountvol, which needs administrator rights")
			exit = command.ExitFailure
		}
	}
	return exit
}

func init() {
	command.Register(mountCommand{})
	command.Register(umountCommand{})
}
