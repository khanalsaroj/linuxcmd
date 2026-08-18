package commands

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// lsofCommand answers the question lsof is reached for most often on a
// developer machine: which process is holding a port. It is built on the
// documented GetExtendedTcpTable/GetExtendedUdpTable interfaces, which
// give the owning PID for every endpoint without any special privilege.
//
// The file half of lsof is deliberately absent. Listing a process's open
// file handles on Windows means enumerating kernel handles through
// NtQuerySystemInformation, an undocumented interface that also needs
// administrator rights and can block on certain handle types. A partial,
// privilege-dependent implementation would be worse than pointing at
// Sysinternals handle.exe, so lsof here reports network endpoints only
// and says so when asked for anything else.
type lsofCommand struct{}

func (lsofCommand) Name() string    { return "lsof" }
func (lsofCommand) Summary() string { return "list open network endpoints and their processes" }

var lsofSpec = parser.Spec{
	{Short: 'i', Long: "internet"},
	{Short: 'n'}, // no host resolution; already the behavior
	{Short: 'P'}, // no port-name resolution; already the behavior
	{Short: 't', Long: "terse"},
	{Short: 'p', HasArg: true},
}

// lsofEndpoint is one row of the merged TCP and UDP tables.
type lsofEndpoint struct {
	Protocol string // "TCP" or "UDP"
	Family   string // "IPv4" or "IPv6"
	Local    string
	Remote   string
	State    string
	PID      uint32
}

func (lsofCommand) Run(ctx *command.Context) int {
	args, filter := splitOptionalArgFlag(ctx.Args, 'i')

	res, err := parser.Parse(args, lsofSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "lsof: %s\n", err)
		return command.ExitUsage
	}
	// A filter may also arrive as a bare operand, e.g. "lsof -i :8080".
	for _, p := range res.Positional {
		if strings.HasPrefix(p, ":") || strings.HasPrefix(p, "@") {
			filter = p
		}
	}

	pidFilter := uint32(0)
	if v, ok := res.Value('p', ""); ok {
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "lsof: invalid process id '%s'\n", v)
			return command.ExitUsage
		}
		pidFilter = uint32(n)
	}

	endpoints, err := networkEndpoints()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "lsof: %s\n", err)
		return command.ExitFailure
	}

	names := processNames()
	terse := res.Bool('t', "terse")

	var rows []lsofEndpoint
	for _, e := range endpoints {
		if pidFilter != 0 && e.PID != pidFilter {
			continue
		}
		if !lsofMatchesFilter(e, filter) {
			continue
		}
		rows = append(rows, e)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].PID != rows[j].PID {
			return rows[i].PID < rows[j].PID
		}
		return rows[i].Local < rows[j].Local
	})

	if terse {
		// -t prints only PIDs, one per line, for piping into kill.
		seen := map[uint32]bool{}
		for _, e := range rows {
			if !seen[e.PID] {
				seen[e.PID] = true
				fmt.Fprintln(ctx.Stdout, e.PID)
			}
		}
		return command.ExitSuccess
	}

	fmt.Fprintf(ctx.Stdout, "%-20s %7s %-5s %-5s %s\n", "COMMAND", "PID", "TYPE", "NODE", "NAME")
	for _, e := range rows {
		name := e.Local
		if e.Remote != "" {
			name += "->" + e.Remote
		}
		if e.State != "" {
			name += " (" + e.State + ")"
		}
		procName := names[e.PID]
		if procName == "" {
			procName = "-"
		}
		fmt.Fprintf(ctx.Stdout, "%-20s %7d %-5s %-5s %s\n", procName, e.PID, e.Family, e.Protocol, name)
	}
	return command.ExitSuccess
}

// splitOptionalArgFlag pulls the value out of a flag that takes an
// optional attached argument, e.g. "-i:8080" or "-i@host". The shared
// parser models flags as either always or never taking a value, and lsof
// -i is the rare case that is genuinely optional, so the token is split
// before parsing rather than complicating the parser for one caller.
func splitOptionalArgFlag(args []string, flag byte) ([]string, string) {
	prefix := "-" + string(flag)
	out := make([]string, 0, len(args))
	value := ""
	for _, a := range args {
		if strings.HasPrefix(a, prefix) && len(a) > len(prefix) {
			value = a[len(prefix):]
			out = append(out, prefix)
			continue
		}
		out = append(out, a)
	}
	return out, value
}

// lsofMatchesFilter applies an -i selector: ":PORT" matches either end,
// "@HOST" matches either address, and an empty filter matches everything.
func lsofMatchesFilter(e lsofEndpoint, filter string) bool {
	switch {
	case filter == "":
		return true
	case strings.HasPrefix(filter, ":"):
		port := filter[1:]
		return strings.HasSuffix(e.Local, ":"+port) || strings.HasSuffix(e.Remote, ":"+port)
	case strings.HasPrefix(filter, "@"):
		host := filter[1:]
		return strings.Contains(e.Local, host) || strings.Contains(e.Remote, host)
	default:
		return strings.Contains(e.Local, filter) || strings.Contains(e.Remote, filter)
	}
}

func processNames() map[uint32]string {
	names := map[uint32]string{}
	procs, err := snapshotProcesses()
	if err != nil {
		return names
	}
	for _, p := range procs {
		names[p.PID] = p.Name
	}
	return names
}

// --- Win32 connection tables ---------------------------------------------

const (
	afInet  = 2
	afInet6 = 23

	tcpTableOwnerPIDAll = 5
	udpTableOwnerPID    = 1

	errorInsufficientBuffer = 122
)

var (
	procGetExtendedTCPTable = iphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUDPTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

// tcpStates maps MIB_TCP_STATE values to the names netstat and lsof use.
var tcpStates = map[uint32]string{
	1: "CLOSED", 2: "LISTEN", 3: "SYN_SENT", 4: "SYN_RCVD",
	5: "ESTABLISHED", 6: "FIN_WAIT1", 7: "FIN_WAIT2", 8: "CLOSE_WAIT",
	9: "CLOSING", 10: "LAST_ACK", 11: "TIME_WAIT", 12: "DELETE_TCB",
}

func networkEndpoints() ([]lsofEndpoint, error) {
	var all []lsofEndpoint

	tcp4, err := tcpEndpoints(afInet)
	if err != nil {
		return nil, err
	}
	all = append(all, tcp4...)
	// IPv6 tables are absent on a machine with the stack disabled, which
	// is not an error worth failing the whole command over.
	if tcp6, err := tcpEndpoints(afInet6); err == nil {
		all = append(all, tcp6...)
	}
	udp4, err := udpEndpoints(afInet)
	if err != nil {
		return nil, err
	}
	all = append(all, udp4...)
	if udp6, err := udpEndpoints(afInet6); err == nil {
		all = append(all, udp6...)
	}
	return all, nil
}

// extendedTable calls one of the GetExtended*Table functions, growing the
// buffer until it fits. The size the API reports can change between the
// sizing call and the real call if a connection opens in between, so the
// call is retried rather than trusted once.
func extendedTable(proc *syscall.LazyProc, family uint32, tableClass uintptr) ([]byte, error) {
	var size uint32
	for attempt := 0; attempt < 5; attempt++ {
		var buf []byte
		var ptr uintptr
		if size > 0 {
			buf = make([]byte, size)
			ptr = uintptr(unsafe.Pointer(&buf[0]))
		}
		ret, _, _ := proc.Call(
			ptr,
			uintptr(unsafe.Pointer(&size)),
			0, // unsorted; we sort for display ourselves
			uintptr(family),
			tableClass,
			0,
		)
		switch ret {
		case 0:
			return buf, nil
		case errorInsufficientBuffer:
			continue
		default:
			return nil, fmt.Errorf("cannot read connection table (error %d)", ret)
		}
	}
	return nil, fmt.Errorf("connection table kept changing size while being read")
}

func tcpEndpoints(family uint32) ([]lsofEndpoint, error) {
	buf, err := extendedTable(procGetExtendedTCPTable, family, tcpTableOwnerPIDAll)
	if err != nil || len(buf) < 4 {
		return nil, err
	}
	count := binary.LittleEndian.Uint32(buf[0:4])
	rows := buf[4:]

	var out []lsofEndpoint
	if family == afInet {
		const rowSize = 24
		for i := uint32(0); i < count && int((i+1)*rowSize) <= len(rows); i++ {
			r := rows[i*rowSize:]
			state := readU32(r, 0)
			out = append(out, lsofEndpoint{
				Protocol: "TCP",
				Family:   "IPv4",
				Local:    joinHostPort(ipv4String(readU32(r, 4)), portFromDWORD(readU32(r, 8))),
				Remote:   tcpRemote(state, ipv4String(readU32(r, 12)), portFromDWORD(readU32(r, 16))),
				State:    tcpStates[state],
				PID:      readU32(r, 20),
			})
		}
		return out, nil
	}

	const rowSize6 = 56
	for i := uint32(0); i < count && int((i+1)*rowSize6) <= len(rows); i++ {
		r := rows[i*rowSize6:]
		state := readU32(r, 48)
		out = append(out, lsofEndpoint{
			Protocol: "TCP",
			Family:   "IPv6",
			Local:    joinHostPort(ipv6String(r[0:16]), portFromDWORD(readU32(r, 20))),
			Remote:   tcpRemote(state, ipv6String(r[24:40]), portFromDWORD(readU32(r, 44))),
			State:    tcpStates[state],
			PID:      readU32(r, 52),
		})
	}
	return out, nil
}

func udpEndpoints(family uint32) ([]lsofEndpoint, error) {
	buf, err := extendedTable(procGetExtendedUDPTable, family, udpTableOwnerPID)
	if err != nil || len(buf) < 4 {
		return nil, err
	}
	count := binary.LittleEndian.Uint32(buf[0:4])
	rows := buf[4:]

	var out []lsofEndpoint
	if family == afInet {
		const rowSize = 12
		for i := uint32(0); i < count && int((i+1)*rowSize) <= len(rows); i++ {
			r := rows[i*rowSize:]
			out = append(out, lsofEndpoint{
				Protocol: "UDP",
				Family:   "IPv4",
				Local:    joinHostPort(ipv4String(readU32(r, 0)), portFromDWORD(readU32(r, 4))),
				PID:      readU32(r, 8),
			})
		}
		return out, nil
	}

	const rowSize6 = 28
	for i := uint32(0); i < count && int((i+1)*rowSize6) <= len(rows); i++ {
		r := rows[i*rowSize6:]
		out = append(out, lsofEndpoint{
			Protocol: "UDP",
			Family:   "IPv6",
			Local:    joinHostPort(ipv6String(r[0:16]), portFromDWORD(readU32(r, 20))),
			PID:      readU32(r, 24),
		})
	}
	return out, nil
}

// tcpRemote suppresses the remote address for a listening socket, where
// Windows reports a meaningless 0.0.0.0:0.
func tcpRemote(state uint32, host string, port uint16) string {
	if tcpStates[state] == "LISTEN" {
		return ""
	}
	return joinHostPort(host, port)
}

// readU32 reads a little-endian DWORD out of a MIB table row.
// encoding/binary is used rather than a pointer cast so the reads
// stay correct regardless of alignment, which matters because the
// project ships arm64 and 386 builds alongside amd64.
func readU32(b []byte, off int) uint32 {
	return binary.LittleEndian.Uint32(b[off : off+4])
}

// portFromDWORD extracts a port from the DWORD the MIB tables store it
// in. The value sits in the low 16 bits in network byte order.
func portFromDWORD(v uint32) uint16 {
	p := uint16(v)
	return p<<8 | p>>8
}

func ipv4String(addr uint32) string {
	return net.IPv4(byte(addr), byte(addr>>8), byte(addr>>16), byte(addr>>24)).String()
}

func ipv6String(b []byte) string {
	ip := make(net.IP, 16)
	copy(ip, b)
	return ip.String()
}

func joinHostPort(host string, port uint16) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]:%d", host, port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func init() { command.Register(lsofCommand{}) }
