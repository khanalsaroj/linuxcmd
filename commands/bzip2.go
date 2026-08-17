package commands

import (
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// bzip2Command only decompresses: Go's standard library provides a
// bzip2 reader but no writer, so compression would need an external
// dependency this project doesn't currently take on. Compression
// attempts fail with a clear, documented error rather than doing
// nothing silently.
type bzip2Command struct{}

func (bzip2Command) Name() string    { return "bzip2" }
func (bzip2Command) Summary() string { return "decompress .bz2 files (compression unsupported)" }

var bzip2Spec = parser.Spec{
	{Short: 'd', Long: "decompress"},
	{Short: 'k', Long: "keep"},
}

func (bzip2Command) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, bzip2Spec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "bzip2: %s\n", err)
		return command.ExitUsage
	}
	if !res.Bool('d', "decompress") {
		fmt.Fprintln(ctx.Stderr, "bzip2: compression is not supported in this build (no bzip2 writer without an external dependency); use -d to decompress")
		return command.ExitFailure
	}
	keep := res.Bool('k', "keep")

	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: bzip2 -d FILE.bz2...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range res.Positional {
		if !strings.HasSuffix(arg, ".bz2") {
			fmt.Fprintf(ctx.Stderr, "bzip2: '%s': unknown suffix -- ignored\n", arg)
			exit = command.ExitFailure
			continue
		}
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "bzip2", arg, err)
			exit = command.ExitFailure
			continue
		}
		in, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "bzip2", arg, err)
			exit = command.ExitFailure
			continue
		}
		outPath := strings.TrimSuffix(resolved, ".bz2")
		out, err := os.Create(outPath)
		if err != nil {
			in.Close()
			output.SimpleErrorf(ctx.Stderr, "bzip2", arg, err)
			exit = command.ExitFailure
			continue
		}
		_, copyErr := io.Copy(out, bzip2.NewReader(in))
		in.Close()
		out.Close()
		if copyErr != nil {
			output.SimpleErrorf(ctx.Stderr, "bzip2", arg, copyErr)
			exit = command.ExitFailure
			continue
		}
		if !keep {
			os.Remove(resolved)
		}
	}
	return exit
}

func init() { command.Register(bzip2Command{}) }
