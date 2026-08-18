package commands

import (
	"fmt"
	"sort"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

// lsblkCommand renders Windows storage as the disk/partition tree lsblk
// shows. The mapping is direct for the parts that exist on both systems:
// a physical disk is a disk, a volume is a partition, and a drive letter
// is a mount point. What has no Linux counterpart -- and no Windows
// counterpart in reverse -- is left out rather than faked, so there are
// no MAJ:MIN numbers here.
type lsblkCommand struct{}

func (lsblkCommand) Name() string    { return "lsblk" }
func (lsblkCommand) Summary() string { return "list disks and the volumes on them" }

var lsblkSpec = parser.Spec{
	{Short: 'b', Long: "bytes"},
	{Short: 'f', Long: "fs"},
}

func (lsblkCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, lsblkSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "lsblk: %s\n", err)
		return command.ExitUsage
	}

	vols, err := enumerateVolumes()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "lsblk: %s\n", err)
		return command.ExitFailure
	}

	size := func(n uint64) string {
		if res.Bool('b', "bytes") {
			return fmt.Sprintf("%d", n)
		}
		return output.HumanSize(int64(n))
	}

	if res.Bool('f', "fs") {
		// -f drops the topology and shows filesystem detail instead.
		fmt.Fprintf(ctx.Stdout, "%-8s %-8s %-12s %10s %s\n", "NAME", "FSTYPE", "LABEL", "FSAVAIL", "MOUNTPOINT")
		for _, v := range vols {
			fmt.Fprintf(ctx.Stdout, "%-8s %-8s %-12s %10s %s\n",
				v.Letter, dashIfEmpty(v.FSType), dashIfEmpty(v.Label), size(v.Free), v.Root)
		}
		return command.ExitSuccess
	}

	// Group volumes under the physical disk they live on. Volumes whose
	// disk cannot be determined (network shares, some virtual drives) are
	// collected separately rather than attributed to a disk they may not
	// be on.
	byDisk := map[int][]volumeInfo{}
	for _, v := range vols {
		byDisk[v.DiskNumber] = append(byDisk[v.DiskNumber], v)
	}
	disks := make([]int, 0, len(byDisk))
	for d := range byDisk {
		disks = append(disks, d)
	}
	sort.Ints(disks)

	fmt.Fprintf(ctx.Stdout, "%-12s %3s %10s %3s %-8s %s\n", "NAME", "RM", "SIZE", "RO", "TYPE", "MOUNTPOINT")
	for _, disk := range disks {
		members := byDisk[disk]
		sort.Slice(members, func(i, j int) bool { return members[i].Letter < members[j].Letter })

		if disk >= 0 {
			total := physicalDiskSize(disk)
			if total == 0 {
				// The capacity ioctl can be refused; summing the disk's
				// own volumes is the closest figure available.
				for _, v := range members {
					total += v.Total
				}
			}
			fmt.Fprintf(ctx.Stdout, "%-12s %3s %10s %3s %-8s %s\n",
				fmt.Sprintf("disk%d", disk), "0", size(total), "0", "disk", "")
		}

		for i, v := range members {
			name := v.Letter
			if disk >= 0 {
				name = treeBranch(i == len(members)-1) + v.Letter
			}
			fmt.Fprintf(ctx.Stdout, "%-12s %3s %10s %3s %-8s %s\n",
				name, boolDigit(v.Removable()), size(v.Total), "0", v.TypeName(), v.Root)
		}
	}
	return command.ExitSuccess
}

func treeBranch(last bool) string {
	if last {
		return "`-"
	}
	return "|-"
}

func boolDigit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func init() { command.Register(lsblkCommand{}) }
