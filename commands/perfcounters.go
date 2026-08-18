package commands

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows exposes the rate statistics vmstat and iostat report through
// the Performance Data Helper (PDH) library rather than through anything
// resembling /proc. This file wraps the four PDH calls both commands
// need.
//
// Counters are added with PdhAddEnglishCounter, not PdhAddCounter: the
// counter path "\Memory\Pages/sec" is localized on a non-English
// Windows, and only the English variant accepts the canonical name on
// every system.
//
// Rate counters are meaningless from a single sample -- PDH computes
// them from the delta between two collections -- so a query must be
// collected once, allowed to age, and collected again before its values
// can be read. Both callers already sample on an interval, which fits
// this model exactly.

const (
	pdhFmtDouble = 0x00000200
	// PDH_MORE_DATA, returned when a buffer needs to grow.
	pdhMoreData = 0x800007D2
)

var (
	pdh                    = syscall.NewLazyDLL("pdh.dll")
	procPdhOpenQuery       = pdh.NewProc("PdhOpenQueryW")
	procPdhAddEnglish      = pdh.NewProc("PdhAddEnglishCounterW")
	procPdhCollect         = pdh.NewProc("PdhCollectQueryData")
	procPdhGetFormatted    = pdh.NewProc("PdhGetFormattedCounterValue")
	procPdhCloseQuery      = pdh.NewProc("PdhCloseQuery")
	procPdhExpandWildcards = pdh.NewProc("PdhExpandWildCardPathW")
)

// pdhCounterValue mirrors PDH_FMT_COUNTERVALUE. The union is read as a
// double because every counter here is added with PDH_FMT_DOUBLE.
type pdhCounterValue struct {
	CStatus uint32
	_       uint32 // alignment padding before the union
	Double  float64
}

// perfQuery is a PDH query holding one or more named counters.
type perfQuery struct {
	handle   uintptr
	counters map[string]uintptr
	// collected records whether a first sample has been taken, since
	// rate counters need two.
	collected bool
}

func newPerfQuery() (*perfQuery, error) {
	var handle uintptr
	ret, _, _ := procPdhOpenQuery.Call(0, 0, uintptr(unsafe.Pointer(&handle)))
	if ret != 0 {
		return nil, fmt.Errorf("cannot open performance query (status 0x%X)", ret)
	}
	return &perfQuery{handle: handle, counters: map[string]uintptr{}}, nil
}

// add registers a counter under a short name the caller uses to read it
// back. A counter that does not exist on this system is reported so the
// caller can decide whether to continue without it.
func (q *perfQuery) add(name, path string) error {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var counter uintptr
	ret, _, _ := procPdhAddEnglish.Call(
		q.handle,
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&counter)),
	)
	if ret != 0 {
		return fmt.Errorf("counter %q is not available (status 0x%X)", path, ret)
	}
	q.counters[name] = counter
	return nil
}

// collect takes a sample. Rate counters return a usable value only after
// the second collect, with real time elapsed in between.
func (q *perfQuery) collect() error {
	ret, _, _ := procPdhCollect.Call(q.handle)
	if ret != 0 {
		return fmt.Errorf("cannot collect performance data (status 0x%X)", ret)
	}
	q.collected = true
	return nil
}

// value reads a counter, returning 0 for one that is unavailable or has
// not yet accumulated two samples. Reporting zero rather than failing
// matches how vmstat and iostat behave on their very first line.
func (q *perfQuery) value(name string) float64 {
	counter, ok := q.counters[name]
	if !ok {
		return 0
	}
	var out pdhCounterValue
	ret, _, _ := procPdhGetFormatted.Call(
		counter,
		pdhFmtDouble,
		0,
		uintptr(unsafe.Pointer(&out)),
	)
	if ret != 0 || out.CStatus != 0 {
		return 0
	}
	return out.Double
}

func (q *perfQuery) close() {
	if q.handle != 0 {
		procPdhCloseQuery.Call(q.handle)
		q.handle = 0
	}
}

// expandCounterPath turns a wildcard path such as
// "\PhysicalDisk(*)\Disk Reads/sec" into the concrete per-instance paths
// available on this machine, which is how iostat discovers the disks.
func expandCounterPath(pattern string) ([]string, error) {
	patternPtr, err := syscall.UTF16PtrFromString(pattern)
	if err != nil {
		return nil, err
	}

	size := uint32(0)
	for attempt := 0; attempt < 3; attempt++ {
		var buf []uint16
		var bufPtr uintptr
		if size > 0 {
			buf = make([]uint16, size)
			bufPtr = uintptr(unsafe.Pointer(&buf[0]))
		}
		ret, _, _ := procPdhExpandWildcards.Call(
			0, // search the local machine
			uintptr(unsafe.Pointer(patternPtr)),
			bufPtr,
			uintptr(unsafe.Pointer(&size)),
			0,
		)
		switch {
		case ret == 0 && buf != nil:
			return splitNulSeparated(buf[:size]), nil
		case ret == pdhMoreData || (ret == 0 && buf == nil):
			// First pass only sizes the buffer; go around again.
			continue
		default:
			return nil, fmt.Errorf("cannot expand counter path %q (status 0x%X)", pattern, ret)
		}
	}
	return nil, fmt.Errorf("counter path %q kept changing while being read", pattern)
}
