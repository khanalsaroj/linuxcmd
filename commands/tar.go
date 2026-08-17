package commands

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type tarCommand struct{}

func (tarCommand) Name() string    { return "tar" }
func (tarCommand) Summary() string { return "create or extract tar archives" }

var tarSpec = parser.Spec{
	{Short: 'c'},
	{Short: 'x'},
	{Short: 't'},
	{Short: 'v', Long: "verbose"},
	{Short: 'z', Long: "gzip"},
	{Short: 'f', Long: "file", HasArg: true},
}

func (tarCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, tarSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", err)
		return command.ExitUsage
	}

	create := res.Bool('c', "")
	extract := res.Bool('x', "")
	list := res.Bool('t', "")
	verbose := res.Bool('v', "verbose")
	useGzip := res.Bool('z', "gzip")
	archivePath, ok := res.Value('f', "file")
	if !ok {
		fmt.Fprintln(ctx.Stderr, "tar: -f archive is required")
		return command.ExitUsage
	}

	resolved, err := paths.Resolve(archivePath)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	switch {
	case create:
		return tarCreate(ctx, resolved, res.Positional, useGzip, verbose)
	case extract:
		return tarExtract(ctx, resolved, useGzip, verbose, true)
	case list:
		return tarExtract(ctx, resolved, useGzip, verbose, false)
	default:
		fmt.Fprintln(ctx.Stderr, "tar: specify one of -c, -x, or -t")
		return command.ExitUsage
	}
}

func tarCreate(ctx *command.Context, archivePath string, sources []string, useGzip, verbose bool) int {
	if len(sources) == 0 {
		fmt.Fprintln(ctx.Stderr, "tar: no files or directories specified")
		return command.ExitUsage
	}
	out, err := os.Create(archivePath)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer out.Close()

	var w io.Writer = out
	var gz *gzip.Writer
	if useGzip {
		gz = gzip.NewWriter(out)
		w = gz
	}
	tw := tar.NewWriter(w)

	exit := command.ExitSuccess
	for _, src := range paths.ExpandGlobs(sources) {
		srcResolved, err := paths.Resolve(src)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "tar", src, err)
			exit = command.ExitFailure
			continue
		}
		if _, err := os.Stat(srcResolved); err != nil {
			output.SimpleErrorf(ctx.Stderr, "tar", src, err)
			exit = command.ExitFailure
			continue
		}
		base := filepath.Dir(srcResolved)
		walkErr := filepath.Walk(srcResolved, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(base, p)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(rel)
			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}
			hdr.Name = name
			if fi.IsDir() {
				hdr.Name += "/"
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if verbose {
				fmt.Fprintln(ctx.Stdout, hdr.Name)
			}
			if fi.IsDir() {
				return nil
			}
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		})
		if walkErr != nil {
			output.SimpleErrorf(ctx.Stderr, "tar", src, walkErr)
			exit = command.ExitFailure
		}
	}

	if err := tw.Close(); err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
		exit = command.ExitFailure
	}
	if gz != nil {
		if err := gz.Close(); err != nil {
			fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
			exit = command.ExitFailure
		}
	}
	return exit
}

func tarExtract(ctx *command.Context, archivePath string, useGzip, verbose, write bool) int {
	f, err := os.Open(archivePath)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer f.Close()

	var r io.Reader = f
	if useGzip {
		gz, err := gzip.NewReader(f)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		defer gz.Close()
		r = gz
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	tr := tar.NewReader(r)
	exit := command.ExitSuccess
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "tar: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}

		if !write {
			fmt.Fprintln(ctx.Stdout, hdr.Name)
			continue
		}

		target, ok := safeJoin(cwd, hdr.Name)
		if !ok {
			fmt.Fprintf(ctx.Stderr, "tar: refusing to extract '%s' outside destination\n", hdr.Name)
			exit = command.ExitFailure
			continue
		}
		if verbose {
			fmt.Fprintln(ctx.Stdout, hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				output.SimpleErrorf(ctx.Stderr, "tar", hdr.Name, err)
				exit = command.ExitFailure
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				output.SimpleErrorf(ctx.Stderr, "tar", hdr.Name, err)
				exit = command.ExitFailure
				continue
			}
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				output.SimpleErrorf(ctx.Stderr, "tar", hdr.Name, err)
				exit = command.ExitFailure
				continue
			}
			_, copyErr := io.Copy(out, tr)
			out.Close()
			if copyErr != nil {
				output.SimpleErrorf(ctx.Stderr, "tar", hdr.Name, copyErr)
				exit = command.ExitFailure
			}
		}
	}
	return exit
}

func init() { command.Register(tarCommand{}) }
