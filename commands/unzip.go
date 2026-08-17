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

type unzipCommand struct{}

func (unzipCommand) Name() string    { return "unzip" }
func (unzipCommand) Summary() string { return "extract a ZIP archive" }

var unzipSpec = parser.Spec{
	{Short: 'd', HasArg: true},
}

func (unzipCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, unzipSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "unzip: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) < 1 {
		fmt.Fprintln(ctx.Stderr, "usage: unzip archive.zip [-d dir]")
		return command.ExitUsage
	}

	archivePath, err := paths.Resolve(res.Positional[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "unzip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	destDir := "."
	if v, ok := res.Value('d', ""); ok {
		destDir = v
	}
	destResolved, err := paths.Resolve(destDir)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "unzip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	if err := os.MkdirAll(destResolved, 0755); err != nil {
		fmt.Fprintf(ctx.Stderr, "unzip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "unzip: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer r.Close()

	exit := command.ExitSuccess
	for _, f := range r.File {
		target, ok := safeJoin(destResolved, f.Name)
		if !ok {
			fmt.Fprintf(ctx.Stderr, "unzip: refusing to extract '%s' outside destination\n", f.Name)
			exit = command.ExitFailure
			continue
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				output.SimpleErrorf(ctx.Stderr, "unzip", f.Name, err)
				exit = command.ExitFailure
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			output.SimpleErrorf(ctx.Stderr, "unzip", f.Name, err)
			exit = command.ExitFailure
			continue
		}
		rc, err := f.Open()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "unzip", f.Name, err)
			exit = command.ExitFailure
			continue
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			output.SimpleErrorf(ctx.Stderr, "unzip", f.Name, err)
			exit = command.ExitFailure
			continue
		}
		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()
		if copyErr != nil {
			output.SimpleErrorf(ctx.Stderr, "unzip", f.Name, copyErr)
			exit = command.ExitFailure
		} else {
			fmt.Fprintln(ctx.Stdout, target)
		}
	}
	return exit
}

// safeJoin joins base and name, refusing paths that would escape base via
// ".." components or an absolute/rooted name (zip-slip protection).
func safeJoin(base, name string) (string, bool) {
	clean := filepath.Clean(strings.ReplaceAll(name, "\\", "/"))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", false
	}
	target := filepath.Join(base, clean)
	baseWithSep := base + string(filepath.Separator)
	if !strings.HasPrefix(target+string(filepath.Separator), baseWithSep) && target != base {
		return "", false
	}
	return target, true
}

func init() { command.Register(unzipCommand{}) }
