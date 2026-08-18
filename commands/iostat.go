package commands

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"syscall"
	"time"
	"unsafe"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/version"
)

// iostatCommand reports CPU utilization and per-disk I/O. CPU figures
// come from performance counters; disk figures come from the storage
// driver's own statistics via IOCTL_DISK_PERFORMANCE, sampled twice so
// that rates are real measurements rather than estimates.
//
// The %iowait and %steal columns are always zero. That is not a gap in
// the implementation: Windows does not account for time blocked on I/O
// as a distinct processor state (it is counted inside idle), and it does
// not expose hypervisor steal time through these counters.
type iostatCommand struct{}

func (iostatCommand) Name() string    { return "iostat" }
func (iostatCommand) Summary() string { return "report CPU and disk I/O statistics" }

var iostatSpec = parser.Spec{
	{Short: 'c', Long: "cpu"},  // CPU report only
	{Short: 'd', Long: "disk"}, // device report only
	{Short: 'k', Long: "kilo"}, // kilobytes (the default)
	{Short: 'm', Long: "mega"}, // megabytes
	{Short: 'x', Long: "extend"},
}

// diskPerformance mirrors the Win32 DISK_PERFORMANCE struct returned by
// IOCTL_DISK_PERFORMANCE.
type diskPerformance struct {
	BytesRead           int64
	BytesWritten        int64
	ReadTime            int64
	WriteTime           int64
	IdleTime            int64
	ReadCount           uint32
	WriteCount          uint32
	QueueDepth          uint32
	SplitCount          uint32
	QueryTime           int64
	StorageDeviceNumber uint32
	StorageManagerName  [8]uint16
}

const ioctlDiskPerformance = 0x70020

func (iostatCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, iostatSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "iostat: %s\n", err)
		return command.ExitUsage
	}

	delay, count, err := parseIntervalArgs(res.Positional)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "iostat: %s\n", err)
		return command.ExitUsage
	}

	unit, unitName := float64(1024), "kB"
	if res.Bool('m', "mega") {
		unit, unitName = 1024*1024, "MB"
	}

	cpuOnly := res.Bool('c', "cpu")
	diskOnly := res.Bool('d', "disk")
	showCPU := !diskOnly
	showDisk := !cpuOnly

	q, err := newPerfQuery()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "iostat: %s\n", err)
		return command.ExitFailure
	}
	defer q.close()
	_ = q.add("user", `\Processor(_Total)\% User Time`)
	_ = q.add("system", `\Processor(_Total)\% Privileged Time`)
	_ = q.add("idle", `\Processor(_Total)\% Idle Time`)
	if err := q.collect(); err != nil {
		fmt.Fprintf(ctx.Stderr, "iostat: %s\n", err)
		return command.ExitFailure
	}

	disks := physicalDiskNumbers()
	previous := sampleDisks(disks)
	previousAt := time.Now()

	fmt.Fprintf(ctx.Stdout, "linuxcmd %s (%s) \t%s \t_%s_\t(%d CPU)\n\n",
		version.String(), hostnameOrUnknown(), time.Now().Format("01/02/2006"),
		runtime.GOARCH, runtime.NumCPU())

	for i := 0; count == 0 || i < count; i++ {
		wait := delay
		if i == 0 && delay == 0 {
			// A first sample needs a moment of elapsed time before any
			// rate can be computed from it.
			wait = 250 * time.Millisecond
		}
		if wait > 0 {
			time.Sleep(wait)
		}
		if err := q.collect(); err != nil {
			fmt.Fprintf(ctx.Stderr, "iostat: %s\n", err)
			return command.ExitFailure
		}

		current := sampleDisks(disks)
		now := time.Now()
		elapsed := now.Sub(previousAt).Seconds()
		if elapsed <= 0 {
			elapsed = 1
		}

		if showCPU {
			// Written with io.WriteString because the header is full of
			// literal % signs, which any fmt printer reads as format
			// directives.
			io.WriteString(ctx.Stdout, "avg-cpu:  %user   %nice %system %iowait  %steal   %idle\n")
			fmt.Fprintf(ctx.Stdout, "        %7.2f %7.2f %7.2f %7.2f %7.2f %7.2f\n\n",
				q.value("user"), 0.0, q.value("system"), 0.0, 0.0, q.value("idle"))
		}

		if showDisk {
			fmt.Fprintf(ctx.Stdout, "%-12s %8s %12s %12s %10s %10s\n",
				"Device", "tps", unitName+"_read/s", unitName+"_wrtn/s", unitName+"_read", unitName+"_wrtn")
			for _, n := range disks {
				prev, hadPrev := previous[n]
				cur, ok := current[n]
				if !ok {
					continue
				}
				var tps, readRate, writeRate float64
				if hadPrev {
					ops := float64(cur.ReadCount+cur.WriteCount) - float64(prev.ReadCount+prev.WriteCount)
					tps = ops / elapsed
					readRate = float64(cur.BytesRead-prev.BytesRead) / unit / elapsed
					writeRate = float64(cur.BytesWritten-prev.BytesWritten) / unit / elapsed
				}
				fmt.Fprintf(ctx.Stdout, "%-12s %8.2f %12.2f %12.2f %10.0f %10.0f\n",
					fmt.Sprintf("disk%d", n), tps, readRate, writeRate,
					float64(cur.BytesRead)/unit, float64(cur.BytesWritten)/unit)
			}
			fmt.Fprintln(ctx.Stdout)
		}

		previous, previousAt = current, now
		if delay == 0 {
			break
		}
	}
	return command.ExitSuccess
}

// physicalDiskNumbers returns the disks that currently back a mounted
// volume, which is the same set lsblk reports so the two agree.
func physicalDiskNumbers() []int {
	seen := map[int]bool{}
	vols, err := enumerateVolumes()
	if err != nil {
		return nil
	}
	for _, v := range vols {
		if v.DiskNumber >= 0 {
			seen[v.DiskNumber] = true
		}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

func sampleDisks(disks []int) map[int]diskPerformance {
	out := map[int]diskPerformance{}
	for _, n := range disks {
		if perf, ok := readDiskPerformance(n); ok {
			out[n] = perf
		}
	}
	return out
}

// readDiskPerformance queries the storage driver's cumulative counters
// for one physical disk. The handle is opened with no access rights,
// which is enough for this ioctl and needs no elevation.
func readDiskPerformance(diskNumber int) (diskPerformance, bool) {
	var perf diskPerformance

	path, err := syscall.UTF16PtrFromString(fmt.Sprintf(`\\.\PhysicalDrive%d`, diskNumber))
	if err != nil {
		return perf, false
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
		return perf, false
	}
	defer syscall.CloseHandle(handle)

	var returned uint32
	if err := syscall.DeviceIoControl(
		handle,
		ioctlDiskPerformance,
		nil, 0,
		(*byte)(unsafe.Pointer(&perf)), uint32(unsafe.Sizeof(perf)),
		&returned, nil,
	); err != nil {
		return perf, false
	}
	return perf, true
}

func hostnameOrUnknown() string {
	name, err := syscall.ComputerName()
	if err != nil {
		return "unknown"
	}
	return name
}

func init() { command.Register(iostatCommand{}) }
