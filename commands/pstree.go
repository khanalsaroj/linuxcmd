package commands

import (
	"fmt"
	"sort"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type pstreeCommand struct{}

func (pstreeCommand) Name() string    { return "pstree" }
func (pstreeCommand) Summary() string { return "print a process tree" }

var pstreeSpec = parser.Spec{
	{Short: 'p'},
}

func (pstreeCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, pstreeSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pstree: %s\n", err)
		return command.ExitUsage
	}
	showPID := res.Bool('p', "")

	procs, err := snapshotProcesses()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "pstree: %s\n", err)
		return command.ExitFailure
	}

	byPID := make(map[uint32]procInfo, len(procs))
	children := make(map[uint32][]uint32)
	for _, p := range procs {
		byPID[p.PID] = p
	}
	for _, p := range procs {
		if _, ok := byPID[p.PPID]; ok && p.PPID != p.PID {
			children[p.PPID] = append(children[p.PPID], p.PID)
		}
	}
	for pid := range children {
		sort.Slice(children[pid], func(i, j int) bool { return children[pid][i] < children[pid][j] })
	}

	hasParent := make(map[uint32]bool, len(procs))
	for _, kids := range children {
		for _, pid := range kids {
			hasParent[pid] = true
		}
	}

	var roots []uint32
	for _, p := range procs {
		if !hasParent[p.PID] {
			roots = append(roots, p.PID)
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })

	var printNode func(pid uint32, prefix string)
	printNode = func(pid uint32, prefix string) {
		p := byPID[pid]
		if showPID {
			fmt.Fprintf(ctx.Stdout, "%s%s(%d)\n", prefix, p.Name, p.PID)
		} else {
			fmt.Fprintf(ctx.Stdout, "%s%s\n", prefix, p.Name)
		}
		for _, child := range children[pid] {
			printNode(child, prefix+"  ")
		}
	}
	for _, root := range roots {
		printNode(root, "")
	}
	return command.ExitSuccess
}

func init() { command.Register(pstreeCommand{}) }
