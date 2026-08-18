package commands

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

// ifconfigCommand renders Windows adapter configuration in net-tools
// ifconfig layout. The project already has "ip addr", but ifconfig is
// what most people still type, and the two formats are different enough
// that output pasted from one does not read as the other.
//
// It is read-only. Reconfiguring an interface on Windows goes through
// netsh or the Set-NetIPAddress cmdlets, needs administrator rights, and
// has no faithful ifconfig spelling, so the write half is left out
// rather than half-implemented.
type ifconfigCommand struct{}

func (ifconfigCommand) Name() string    { return "ifconfig" }
func (ifconfigCommand) Summary() string { return "display network interface configuration" }

var ifconfigSpec = parser.Spec{
	{Short: 'a', Long: "all"},
	{Short: 's', Long: "short"},
}

func (ifconfigCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, ifconfigSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ifconfig: %s\n", err)
		return command.ExitUsage
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "ifconfig: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	// A named operand selects a single adapter; otherwise ifconfig shows
	// the interfaces that are up, and -a shows every one.
	wanted := ""
	if len(res.Positional) > 0 {
		wanted = res.Positional[0]
	}
	showAll := res.Bool('a', "all") || wanted != ""

	if res.Bool('s', "short") {
		fmt.Fprintf(ctx.Stdout, "%-24s %6s %10s %10s %10s %10s\n",
			"Iface", "MTU", "RX-OK", "RX-ERR", "TX-OK", "TX-ERR")
	}

	matched := false
	for _, iface := range ifaces {
		if wanted != "" && !strings.EqualFold(iface.Name, wanted) {
			continue
		}
		if !showAll && iface.Flags&net.FlagUp == 0 {
			continue
		}
		matched = true
		if res.Bool('s', "short") {
			ifconfigShortLine(ctx, iface)
		} else {
			ifconfigDetail(ctx, iface)
		}
	}

	if wanted != "" && !matched {
		fmt.Fprintf(ctx.Stderr, "ifconfig: interface %s does not exist\n", wanted)
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func ifconfigShortLine(ctx *command.Context, iface net.Interface) {
	stats, _ := interfaceStats(uint32(iface.Index))
	fmt.Fprintf(ctx.Stdout, "%-24s %6d %10d %10d %10d %10d\n",
		iface.Name, iface.MTU, stats.InUcastPkts, stats.InErrors, stats.OutUcastPkts, stats.OutErrors)
}

func ifconfigDetail(ctx *command.Context, iface net.Interface) {
	fmt.Fprintf(ctx.Stdout, "%s: flags=<%s>  mtu %d\n", iface.Name, ifconfigFlags(iface), iface.MTU)

	addrs, _ := iface.Addrs()
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if v4 := ipNet.IP.To4(); v4 != nil {
			mask := net.IP(ipNet.Mask).String()
			fmt.Fprintf(ctx.Stdout, "        inet %s  netmask %s", v4, mask)
			if bcast := broadcastAddr(v4, ipNet.Mask); bcast != nil {
				fmt.Fprintf(ctx.Stdout, "  broadcast %s", bcast)
			}
			fmt.Fprintln(ctx.Stdout)
			continue
		}
		prefix, _ := ipNet.Mask.Size()
		scope := "global"
		if ipNet.IP.IsLinkLocalUnicast() {
			scope = "link"
		}
		fmt.Fprintf(ctx.Stdout, "        inet6 %s  prefixlen %d  scopeid <%s>\n", ipNet.IP, prefix, scope)
	}

	if len(iface.HardwareAddr) > 0 {
		fmt.Fprintf(ctx.Stdout, "        ether %s\n", iface.HardwareAddr)
	}

	if stats, err := interfaceStats(uint32(iface.Index)); err == nil {
		fmt.Fprintf(ctx.Stdout, "        RX packets %d  bytes %d\n", stats.InUcastPkts+stats.InNUcastPkts, stats.InOctets)
		fmt.Fprintf(ctx.Stdout, "        RX errors %d  dropped %d\n", stats.InErrors, stats.InDiscards)
		fmt.Fprintf(ctx.Stdout, "        TX packets %d  bytes %d\n", stats.OutUcastPkts+stats.OutNUcastPkts, stats.OutOctets)
		fmt.Fprintf(ctx.Stdout, "        TX errors %d  dropped %d\n", stats.OutErrors, stats.OutDiscards)
	}
	fmt.Fprintln(ctx.Stdout)
}

func ifconfigFlags(iface net.Interface) string {
	var flags []string
	if iface.Flags&net.FlagUp != 0 {
		flags = append(flags, "UP", "RUNNING")
	}
	if iface.Flags&net.FlagBroadcast != 0 {
		flags = append(flags, "BROADCAST")
	}
	if iface.Flags&net.FlagLoopback != 0 {
		flags = append(flags, "LOOPBACK")
	}
	if iface.Flags&net.FlagPointToPoint != 0 {
		flags = append(flags, "POINTOPOINT")
	}
	if iface.Flags&net.FlagMulticast != 0 {
		flags = append(flags, "MULTICAST")
	}
	if len(flags) == 0 {
		return "DOWN"
	}
	return strings.Join(flags, ",")
}

// broadcastAddr derives the IPv4 broadcast address from an address and
// its mask. Windows does not report one directly.
func broadcastAddr(ip net.IP, mask net.IPMask) net.IP {
	v4 := ip.To4()
	if v4 == nil || len(mask) != net.IPv4len {
		return nil
	}
	out := make(net.IP, net.IPv4len)
	for i := range out {
		out[i] = v4[i] | ^mask[i]
	}
	return out
}

// mibIfRow mirrors the Win32 MIB_IFROW struct used by GetIfEntry. The
// older IPv4-era API is enough here and has a far simpler layout than
// MIB_IF_ROW2; the tradeoff is that its counters are 32-bit and wrap on
// a busy interface, which ifconfig's own counters historically did too.
type mibIfRow struct {
	Name            [256]uint16
	Index           uint32
	Type            uint32
	Mtu             uint32
	Speed           uint32
	PhysAddrLen     uint32
	PhysAddr        [8]byte
	AdminStatus     uint32
	OperStatus      uint32
	LastChange      uint32
	InOctets        uint32
	InUcastPkts     uint32
	InNUcastPkts    uint32
	InDiscards      uint32
	InErrors        uint32
	InUnknownProtos uint32
	OutOctets       uint32
	OutUcastPkts    uint32
	OutNUcastPkts   uint32
	OutDiscards     uint32
	OutErrors       uint32
	OutQLen         uint32
	DescrLen        uint32
	Descr           [256]byte
}

var (
	iphlpapi       = syscall.NewLazyDLL("iphlpapi.dll")
	procGetIfEntry = iphlpapi.NewProc("GetIfEntry")
)

func interfaceStats(index uint32) (mibIfRow, error) {
	var row mibIfRow
	row.Index = index
	ret, _, _ := procGetIfEntry.Call(uintptr(unsafe.Pointer(&row)))
	if ret != 0 {
		return row, fmt.Errorf("GetIfEntry failed for interface %d (error %d)", index, ret)
	}
	return row, nil
}

func init() { command.Register(ifconfigCommand{}) }
