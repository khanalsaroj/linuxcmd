package commands

import (
	"fmt"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
)

type diffCommand struct{}

func (diffCommand) Name() string    { return "diff" }
func (diffCommand) Summary() string { return "compare two files line by line" }

var diffSpec = parser.Spec{
	{Short: 'u', Long: "unified"},
}

// diffOp is one line of an edit script turning lines1 into lines2.
type diffOp struct {
	kind byte // ' ' (equal), '-' (delete from file1), '+' (insert from file2)
	text string
}

// diffLines computes a minimal edit script between a and b using a
// classic O(n*m) LCS dynamic-programming table. That's fine for the
// file sizes this command is meant for; a Myers-style O(ND) algorithm
// would be needed for very large inputs.
func diffLines(a, b []string) []diffOp {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var ops []diffOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, diffOp{' ', a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, diffOp{'-', a[i]})
			i++
		default:
			ops = append(ops, diffOp{'+', b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{'-', a[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{'+', b[j]})
	}
	return ops
}

func (diffCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, diffSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "diff: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) != 2 {
		fmt.Fprintln(ctx.Stderr, "usage: diff [-u] FILE1 FILE2")
		return command.ExitUsage
	}
	unified := res.Bool('u', "unified")

	lines1, err := readAllLines(res.Positional[0], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "diff", res.Positional[0], err)
		return command.ExitFailure
	}
	lines2, err := readAllLines(res.Positional[1], ctx)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, "diff", res.Positional[1], err)
		return command.ExitFailure
	}

	ops := diffLines(lines1, lines2)

	changed := false
	for _, op := range ops {
		if op.kind != ' ' {
			changed = true
			break
		}
	}
	if !changed {
		return command.ExitSuccess
	}

	if unified {
		fmt.Fprintf(ctx.Stdout, "--- %s\n", res.Positional[0])
		fmt.Fprintf(ctx.Stdout, "+++ %s\n", res.Positional[1])
		for _, op := range ops {
			switch op.kind {
			case ' ':
				fmt.Fprintf(ctx.Stdout, " %s\n", op.text)
			case '-':
				fmt.Fprintf(ctx.Stdout, "-%s\n", op.text)
			case '+':
				fmt.Fprintf(ctx.Stdout, "+%s\n", op.text)
			}
		}
	} else {
		for _, op := range ops {
			switch op.kind {
			case '-':
				fmt.Fprintf(ctx.Stdout, "< %s\n", op.text)
			case '+':
				fmt.Fprintf(ctx.Stdout, "> %s\n", op.text)
			}
		}
	}
	return command.ExitFailure
}

func init() { command.Register(diffCommand{}) }
