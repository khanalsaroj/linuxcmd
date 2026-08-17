package commands

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type zipCommand struct{}

func (zipCommand) Name() string    { return "zip" }
func (zipCommand) Summary() string { return "create a ZIP archive" }

var zipSpec = parser.Spec{
	{Short: 'r'},
}

func (zipCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, zipSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "zip: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) < 2 {
		fmt.Fprintln(ctx.Stderr, "usage: zip [-r] archive.zip file...")
		return command.ExitUsage
	}
	recursive := res.Bool('r', "")

	archivePath, err := paths.Resolve(res.Positional[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "zip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	out, err := os.Create(archivePath)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "zip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	exit := command.ExitSuccess

	addFile := func(diskPath, archiveName string) error {
		f, err := os.Open(diskPath)
		if err != nil {
			return err
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(archiveName)
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	}

	for _, src := range paths.ExpandGlobs(res.Positional[1:len(res.Positional)]) {
		resolved, err := paths.Resolve(src)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "zip", src, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "zip", src, err)
			exit = command.ExitFailure
			continue
		}
		if info.IsDir() {
			if !recursive {
				fmt.Fprintf(ctx.Stderr, "zip: '%s' is a directory (use -r)\n", src)
				exit = command.ExitFailure
				continue
			}
			base := filepath.Dir(resolved)
			walkErr := filepath.Walk(resolved, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return nil
				}
				rel, err := filepath.Rel(base, p)
				if err != nil {
					return err
				}
				return addFile(p, strings.ReplaceAll(rel, "\\", "/"))
			})
			if walkErr != nil {
				output.SimpleErrorf(ctx.Stderr, "zip", src, walkErr)
				exit = command.ExitFailure
			}
			continue
		}
		if err := addFile(resolved, filepath.Base(resolved)); err != nil {
			output.SimpleErrorf(ctx.Stderr, "zip", src, err)
			exit = command.ExitFailure
		}
	}

	if err := zw.Close(); err != nil {
		fmt.Fprintf(ctx.Stderr, "zip: %s\n", output.LinuxErrorText(err))
		exit = command.ExitFailure
	}
	return exit
}

func init() { command.Register(zipCommand{}) }
