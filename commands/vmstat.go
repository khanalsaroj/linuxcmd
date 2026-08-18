package commands

import (
	"fmt"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

// vmstatCommand reports memory, paging and CPU activity on vmstat's
// interval model. Figures come from GlobalMemoryStatusEx for the memory
// columns and from performance counters for the rate columns.
//
// Three columns have no Windows counterpart and are reported as zero
// rather than filled with a plausible-looking number: "b" (processes
// blocked on I/O, which Windows does not account for separately), "buff"
// (Windows has no buffer cache distinct from the system cache), and "st"
// (steal time, a virtualization concept the Windows counters do not
// expose). Saying zero and documenting why is better than inventing a
// value that would be read as real.
type vmstatCommand struct{}

func (vmstatCommand) Name() string    { return "vmstat" }
func (vmstatCommand) Summary() string { return "report memory, paging, and CPU activity" }

var vmstatSpec = parser.Spec{
	{Short: 'S', HasArg: true}, // unit: k, K, m, M
	{Short: 'n', Long: "one-header"},
}

func (vmstatCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, vmstatSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
		return command.ExitUsage
	}

	delay, count, err := parseIntervalArgs(res.Positional)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
		return command.ExitUsage
	}

	unit := uint64(1024) // vmstat reports kilobytes by default
	if v, ok := res.Value('S', ""); ok {
		switch v {
		case "k", "K":
			unit = 1024
		case "m", "M":
			unit = 1024 * 1024
		default:
			fmt.Fprintf(ctx.Stderr, "vmstat: invalid unit '%s'; use k or m\n", v)
			return command.ExitUsage
		}
	}

	q, err := newPerfQuery()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
		return command.ExitFailure
	}
	defer q.close()

	// A counter missing on this system is not fatal: its column reports
	// zero, which is the same treatment the columns Windows has no
	// concept of already get.
	_ = q.add("runqueue", `\System\Processor Queue Length`)
	_ = q.add("pagein", `\Memory\Pages Input/sec`)
	_ = q.add("pageout", `\Memory\Pages Output/sec`)
	_ = q.add("diskread", `\PhysicalDisk(_Total)\Disk Reads/sec`)
	_ = q.add("diskwrite", `\PhysicalDisk(_Total)\Disk Writes/sec`)
	_ = q.add("interrupts", `\Processor(_Total)\Interrupts/sec`)
	_ = q.add("switches", `\System\Context Switches/sec`)
	_ = q.add("user", `\Processor(_Total)\% User Time`)
	_ = q.add("system", `\Processor(_Total)\% Privileged Time`)
	_ = q.add("idle", `\Processor(_Total)\% Idle Time`)

	if err := q.collect(); err != nil {
		fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
		return command.ExitFailure
	}

	header := func() {
		fmt.Fprintln(ctx.Stdout, "procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----")
		fmt.Fprintln(ctx.Stdout, " r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st")
	}
	header()

	for i := 0; count == 0 || i < count; i++ {
		// The first line of vmstat is conventionally an average since
		// boot; here it is the interval since the query opened, which is
		// the closest honest equivalent and takes a moment to accumulate.
		wait := delay
		if i == 0 && delay == 0 {
			wait = 250 * time.Millisecond
		}
		if wait > 0 {
			time.Sleep(wait)
		}
		if err := q.collect(); err != nil {
			fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
			return command.ExitFailure
		}

		mem, err := globalMemoryStatus()
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "vmstat: %s\n", err)
			return command.ExitFailure
		}

		// Windows reports commit charge, not a swap device. The part of
		// the commit charge that exceeds physical memory in use is the
		// closest equivalent to Linux's "swpd".
		physUsed := mem.TotalPhys - mem.AvailPhys
		commitUsed := mem.TotalPageFile - mem.AvailPageFile
		swapped := uint64(0)
		if commitUsed > physUsed {
			swapped = commitUsed - physUsed
		}

		fmt.Fprintf(ctx.Stdout, "%2.0f %2d %6d %6d %6d %6d %4.0f %4.0f %5.0f %5.0f %4.0f %4.0f %2.0f %2.0f %2.0f %2d %2d\n",
			q.value("runqueue"),
			0, // no Windows equivalent of processes blocked on I/O
			swapped/unit,
			mem.AvailPhys/unit,
			0, // no buffer cache distinct from the system cache
			systemCacheBytes()/unit,
			q.value("pagein"),
			q.value("pageout"),
			q.value("diskread"),
			q.value("diskwrite"),
			q.value("interrupts"),
			q.value("switches"),
			q.value("user"),
			q.value("system"),
			q.value("idle"),
			0, // iowait is folded into idle on Windows
			0, // steal time is not exposed
		)

		if delay == 0 {
			break
		}
		if !res.Bool('n', "one-header") && (i+1)%22 == 0 {
			header()
		}
	}
	return command.ExitSuccess
}

// parseIntervalArgs reads the "DELAY [COUNT]" operands vmstat, iostat
// and their relatives share. A zero delay means take a single sample.
func parseIntervalArgs(args []string) (time.Duration, int, error) {
	if len(args) == 0 {
		return 0, 1, nil
	}
	seconds, err := strconv.Atoi(args[0])
	if err != nil || seconds < 0 {
		return 0, 0, fmt.Errorf("invalid delay '%s'", args[0])
	}
	count := 0 // repeat forever unless a count is given
	if len(args) > 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 {
			return 0, 0, fmt.Errorf("invalid count '%s'", args[1])
		}
		count = n
	}
	return time.Duration(seconds) * time.Second, count, nil
}

// performanceInformation mirrors the Win32 PERFORMANCE_INFORMATION
// struct. Only SystemCache is read here, but the whole struct must be
// declared for the call to size correctly.
type performanceInformation struct {
	Size              uint32
	CommitTotal       uintptr
	CommitLimit       uintptr
	CommitPeak        uintptr
	PhysicalTotal     uintptr
	PhysicalAvailable uintptr
	SystemCache       uintptr
	KernelTotal       uintptr
	KernelPaged       uintptr
	KernelNonpaged    uintptr
	PageSize          uintptr
	HandleCount       uint32
	ProcessCount      uint32
	ThreadCount       uint32
}

var (
	psapi                  = syscall.NewLazyDLL("psapi.dll")
	procGetPerformanceInfo = psapi.NewProc("GetPerformanceInfo")
)

// systemCacheBytes reports the size of the Windows system file cache,
// which is the closest counterpart to Linux's "cache" column.
func systemCacheBytes() uint64 {
	var info performanceInformation
	info.Size = uint32(unsafe.Sizeof(info))
	ret, _, _ := procGetPerformanceInfo.Call(uintptr(unsafe.Pointer(&info)), uintptr(info.Size))
	if ret == 0 {
		return 0
	}
	return uint64(info.SystemCache) * uint64(info.PageSize)
}

func init() { command.Register(vmstatCommand{}) }
