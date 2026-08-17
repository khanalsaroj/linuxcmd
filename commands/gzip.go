package commands

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type gzipCommand struct{}

func (gzipCommand) Name() string    { return "gzip" }
func (gzipCommand) Summary() string { return "compress or decompress files with gzip" }

var gzipSpec = parser.Spec{
	{Short: 'k', Long: "keep"},
	{Short: 'd', Long: "decompress"},
}

func (gzipCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, gzipSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "gzip: %s\n", err)
		return command.ExitUsage
	}
	keep := res.Bool('k', "keep")
	decompress := res.Bool('d', "decompress")

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: gzip [-d] [-k] FILE...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "gzip", arg, err)
			exit = command.ExitFailure
			continue
		}

		if decompress {
			if err := gunzipFile(resolved); err != nil {
				output.SimpleErrorf(ctx.Stderr, "gzip", arg, err)
				exit = command.ExitFailure
				continue
			}
		} else {
			if err := gzipFile(resolved); err != nil {
				output.SimpleErrorf(ctx.Stderr, "gzip", arg, err)
				exit = command.ExitFailure
				continue
			}
		}
		if !keep {
			os.Remove(resolved)
		}
	}
	return exit
}

func gzipFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(path + ".gz")
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	if _, err := io.Copy(gw, in); err != nil {
		gw.Close()
		return err
	}
	return gw.Close()
}

func gunzipFile(path string) error {
	if !strings.HasSuffix(path, ".gz") {
		return fmt.Errorf("unknown suffix -- ignored")
	}
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	gr, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gr.Close()

	out, err := os.Create(strings.TrimSuffix(path, ".gz"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, gr)
	return err
}

func init() { command.Register(gzipCommand{}) }
