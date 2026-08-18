package commands

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// getfaclCommand shows the real Windows access-control entries on a
// file, rather than the POSIX approximation that ls -l and chmod present
// everywhere else in this project. When a permission question actually
// matters, this is the command that answers it accurately.
//
// setfacl has deliberately not been implemented. Windows ACLs carry
// information POSIX has no room for: entries can deny as well as allow,
// they are order-sensitive, they can be inherited from a parent
// directory, and they distinguish rights that rwx collapses together.
// Translating rwx back into an ACL would have to discard all of that,
// and a lossy write to an access-control list silently destroys security
// configuration. Reading is safe; writing is not, so only reading is
// offered.
type getfaclCommand struct{}

func (getfaclCommand) Name() string    { return "getfacl" }
func (getfaclCommand) Summary() string { return "show Windows access-control entries for a file" }

var getfaclSpec = parser.Spec{
	{Short: 'a', Long: "access"},
	{Short: 'n', Long: "numeric"},
	{Short: 'c', Long: "omit-header"},
	{Short: 'p', Long: "absolute-names"},
}

// Windows security-information selectors.
const (
	ownerSecurityInformation = 0x00000001
	groupSecurityInformation = 0x00000002
	daclSecurityInformation  = 0x00000004
	seFileObject             = 1
)

// ACE types that carry a plain access mask. Object ACEs (types 5 and 6)
// exist only on directory-service objects, never on files.
const (
	accessAllowedAceType = 0
	accessDeniedAceType  = 1
)

// ACE flags worth surfacing.
const (
	objectInheritAce    = 0x01
	containerInheritAce = 0x02
	inheritedAce        = 0x10
)

// File access rights. The generic rights are included because ACEs
// written by older tools still use them.
const (
	fileReadData        = 0x0001
	fileWriteData       = 0x0002
	fileAppendData      = 0x0004
	fileExecute         = 0x0020
	fileDeleteChild     = 0x0040
	fileWriteAttributes = 0x0100
	writeDAC            = 0x00040000
	writeOwner          = 0x00080000
	genericAll          = 0x10000000
	genericExecute      = 0x20000000
	genericWrite        = 0x40000000
	genericRead         = 0x80000000
)

// SID_NAME_USE values that mean "this is a group of some kind".
const (
	sidTypeUser           = 1
	sidTypeGroup          = 2
	sidTypeAlias          = 4
	sidTypeWellKnownGroup = 5
)

var (
	advapi32                 = syscall.NewLazyDLL("advapi32.dll")
	procGetNamedSecurityInfo = advapi32.NewProc("GetNamedSecurityInfoW")
	procGetAce               = advapi32.NewProc("GetAce")
	procLookupAccountSid     = advapi32.NewProc("LookupAccountSidW")
	procConvertSidToString   = advapi32.NewProc("ConvertSidToStringSidW")
	kernel32Acl              = syscall.NewLazyDLL("kernel32.dll")
	procLocalFree            = kernel32Acl.NewProc("LocalFree")
)

// windowsACL mirrors the Win32 ACL header. Individual entries are
// reached with GetAce rather than by walking the buffer by hand, so the
// only field read here is AceCount.
type windowsACL struct {
	AclRevision byte
	Sbz1        byte
	AclSize     uint16
	AceCount    uint16
	Sbz2        uint16
}

// aceHeader mirrors ACE_HEADER, which prefixes every access-control
// entry.
type aceHeader struct {
	AceType  byte
	AceFlags byte
	AceSize  uint16
}

// fileACE is the layout ACCESS_ALLOWED_ACE and ACCESS_DENIED_ACE share:
// a header, an access mask, then the trustee's SID stored inline.
// SidStart is the SID's first word; its address is the SID's address.
type fileACE struct {
	Header   aceHeader
	Mask     uint32
	SidStart uint32
}

// sidRef points at a SID owned by Windows. A SID is variable-length and
// opaque, so it is only ever passed back to the API by address, never
// dereferenced here.
type sidRef = byte

func (getfaclCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, getfaclSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "getfacl: %s\n", err)
		return command.ExitUsage
	}

	operands := paths.ExpandGlobs(res.Positional)
	if len(operands) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: getfacl [-n] [-c] FILE...")
		return command.ExitUsage
	}

	numeric := res.Bool('n', "numeric")
	omitHeader := res.Bool('c', "omit-header")

	exit := command.ExitSuccess
	for i, name := range operands {
		if i > 0 {
			fmt.Fprintln(ctx.Stdout)
		}
		if !printFileACL(ctx, name, numeric, omitHeader) {
			exit = command.ExitFailure
		}
	}
	return exit
}

func printFileACL(ctx *command.Context, name string, numeric, omitHeader bool) bool {
	resolved, err := paths.Resolve(name)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "getfacl", name, err)
		return false
	}

	sd, err := readSecurityDescriptor(resolved)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "getfacl: %s: %s\n", name, err)
		return false
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(sd.descriptor)))

	if !omitHeader {
		fmt.Fprintf(ctx.Stdout, "# file: %s\n", name)
		fmt.Fprintf(ctx.Stdout, "# owner: %s\n", describeSID(sd.owner, numeric))
		fmt.Fprintf(ctx.Stdout, "# group: %s\n", describeSID(sd.group, numeric))
		// Without this line the output reads like a POSIX ACL, which
		// would misrepresent what Windows actually enforces.
		fmt.Fprintln(ctx.Stdout, "# note: Windows ACL; entries are ordered and may deny as well as allow")
	}

	if sd.dacl == nil {
		// A null DACL grants everyone full access; an empty one denies
		// everyone. The difference matters, so it is spelled out.
		fmt.Fprintln(ctx.Stdout, "# no DACL present: full access is granted to everyone")
		return true
	}

	for i := uint16(0); i < sd.dacl.AceCount; i++ {
		// GetAce writes the entry's address into a typed pointer, so no
		// raw address arithmetic is needed to reach it.
		var ace *fileACE
		ret, _, _ := procGetAce.Call(
			uintptr(unsafe.Pointer(sd.dacl)),
			uintptr(i),
			uintptr(unsafe.Pointer(&ace)),
		)
		if ret == 0 || ace == nil {
			continue
		}
		header := ace.Header
		if header.AceType != accessAllowedAceType && header.AceType != accessDeniedAceType {
			// Audit and object ACEs do not describe access for a file.
			continue
		}

		mask := ace.Mask
		// The SID is stored inline immediately after the mask, so the
		// address of SidStart is the address of the SID.
		sid := (*sidRef)(unsafe.Pointer(&ace.SidStart))

		account, isGroup := lookupSID(sid, numeric)
		kind := "user"
		if isGroup {
			kind = "group"
		}

		line := fmt.Sprintf("%s:%s:%s", kind, account, maskToRWX(mask))
		var notes []string
		if header.AceType == accessDeniedAceType {
			notes = append(notes, "deny")
		}
		if header.AceFlags&inheritedAce != 0 {
			notes = append(notes, "inherited")
		}
		if header.AceFlags&(objectInheritAce|containerInheritAce) != 0 {
			notes = append(notes, "inheritable")
		}
		if extra := maskExtras(mask); extra != "" {
			notes = append(notes, extra)
		}
		if len(notes) > 0 {
			line += "\t#" + strings.Join(notes, ",")
		}
		fmt.Fprintln(ctx.Stdout, line)
	}
	return true
}

// securityDescriptor holds the pointers GetNamedSecurityInfo hands back.
// They all point into one allocation that the caller must LocalFree.
// Every field is a typed pointer the API writes into directly, which
// keeps Windows-owned addresses out of Go's uintptr arithmetic.
type securityDescriptor struct {
	owner      *sidRef
	group      *sidRef
	dacl       *windowsACL
	descriptor *byte
}

func readSecurityDescriptor(path string) (securityDescriptor, error) {
	var sd securityDescriptor

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return sd, err
	}
	ret, _, _ := procGetNamedSecurityInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		seFileObject,
		ownerSecurityInformation|groupSecurityInformation|daclSecurityInformation,
		uintptr(unsafe.Pointer(&sd.owner)),
		uintptr(unsafe.Pointer(&sd.group)),
		uintptr(unsafe.Pointer(&sd.dacl)),
		0, // no SACL; reading one needs SeSecurityPrivilege
		uintptr(unsafe.Pointer(&sd.descriptor)),
	)
	if ret != 0 {
		return sd, fmt.Errorf("cannot read security information: %s",
			output.LinuxErrorText(syscall.Errno(ret)))
	}
	return sd, nil
}

// lookupSID resolves a SID to DOMAIN\Name, falling back to the textual
// SID when no account matches, which happens for deleted accounts and
// for SIDs from a domain this machine cannot reach.
func lookupSID(sid *sidRef, numeric bool) (string, bool) {
	sidAddr := uintptr(unsafe.Pointer(sid))

	var nameLen, domainLen uint32
	var use uint32
	// First call sizes the buffers.
	procLookupAccountSid.Call(0, sidAddr, 0, uintptr(unsafe.Pointer(&nameLen)),
		0, uintptr(unsafe.Pointer(&domainLen)), uintptr(unsafe.Pointer(&use)))
	if nameLen == 0 {
		return sidToString(sid), false
	}

	name := make([]uint16, nameLen)
	domain := make([]uint16, domainLen+1)
	ret, _, _ := procLookupAccountSid.Call(
		0, sidAddr,
		uintptr(unsafe.Pointer(&name[0])), uintptr(unsafe.Pointer(&nameLen)),
		uintptr(unsafe.Pointer(&domain[0])), uintptr(unsafe.Pointer(&domainLen)),
		uintptr(unsafe.Pointer(&use)),
	)
	if ret == 0 {
		return sidToString(sid), false
	}

	// The account type is still worth resolving under -n, so that the
	// user:/group: prefix stays accurate even when the SID is printed
	// raw.
	isGroup := use == sidTypeGroup || use == sidTypeAlias || use == sidTypeWellKnownGroup
	if numeric {
		return sidToString(sid), isGroup
	}

	account := syscall.UTF16ToString(name)
	if d := syscall.UTF16ToString(domain); d != "" {
		account = d + `\` + account
	}
	return account, isGroup
}

func sidToString(sid *sidRef) string {
	var str *uint16
	ret, _, _ := procConvertSidToString.Call(
		uintptr(unsafe.Pointer(sid)),
		uintptr(unsafe.Pointer(&str)),
	)
	if ret == 0 || str == nil {
		return "(unknown)"
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(str)))
	return utf16PtrToString(str)
}

// describeSID renders an owner or group SID, tolerating the case where
// the descriptor did not include one.
func describeSID(sid *sidRef, numeric bool) string {
	if sid == nil {
		return "(none)"
	}
	name, _ := lookupSID(sid, numeric)
	return name
}

// maskToRWX renders a Windows access mask in the closest rwx spelling.
// It is an approximation by nature; maskExtras reports the rights that
// do not fit.
func maskToRWX(mask uint32) string {
	var sb strings.Builder
	if mask&(fileReadData|genericRead|genericAll) != 0 {
		sb.WriteByte('r')
	} else {
		sb.WriteByte('-')
	}
	if mask&(fileWriteData|fileAppendData|genericWrite|genericAll) != 0 {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('-')
	}
	if mask&(fileExecute|genericExecute|genericAll) != 0 {
		sb.WriteByte('x')
	} else {
		sb.WriteByte('-')
	}
	return sb.String()
}

// maskExtras names the rights that rwx cannot express, so that an entry
// granting more than it appears to is not silently understated.
func maskExtras(mask uint32) string {
	var extras []string
	if mask&writeDAC != 0 {
		extras = append(extras, "change-permissions")
	}
	if mask&writeOwner != 0 {
		extras = append(extras, "take-ownership")
	}
	if mask&fileDeleteChild != 0 {
		extras = append(extras, "delete-child")
	}
	if mask&fileWriteAttributes != 0 && mask&fileWriteData == 0 {
		extras = append(extras, "write-attributes")
	}
	return strings.Join(extras, ",")
}

// utf16PtrToString decodes a NUL-terminated UTF-16 string that Windows
// allocated and handed back through an out-parameter. Walking it with
// unsafe.Add on a typed pointer keeps the traversal free of the raw
// uintptr arithmetic that would be unsound on non-Go memory.
func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	var units []uint16
	for ptr := unsafe.Pointer(p); ; ptr = unsafe.Add(ptr, unsafe.Sizeof(uint16(0))) {
		u := *(*uint16)(ptr)
		if u == 0 {
			break
		}
		units = append(units, u)
	}
	return syscall.UTF16ToString(units)
}

func init() { command.Register(getfaclCommand{}) }
