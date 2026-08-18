package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// dos2unix and unix2dos convert line endings in place. This is the one
// conversion Windows users need constantly and that no bundled Windows
// tool does cleanly, so unlike most of the commands here it is not about
// muscle memory: a Git checkout with the wrong core.autocrlf, a file off
// a network share, or a heredoc written by a Windows editor all produce
// files that break shell scripts, Dockerfiles and JSON parsers in ways
// that are tedious to diagnose.
//
// Both directions share one implementation because they differ only in
// the target terminator. unix2dos deliberately normalizes to LF first
// so that a file already using CRLF is left alone rather than being
// doubled up into CR CR LF.

type lineEndingCommand struct {
	name    string
	summary string
	// crlf reports whether the target line ending is CRLF.
	crlf bool
}

func (c lineEndingCommand) Name() string    { return c.name }
func (c lineEndingCommand) Summary() string { return c.summary }

var lineEndingSpec = parser.Spec{
	{Short: 'k', Long: "keepdate"},
	{Short: 'q', Long: "quiet"},
	{Short: 'f', Long: "force"},
	{Short: 'n', Long: "newfile"},
	{Short: 'o', Long: "oldfile"},
}

func (c lineEndingCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, lineEndingSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, err)
		return command.ExitUsage
	}
	quiet := res.Bool('q', "quiet")
	keepDate := res.Bool('k', "keepdate")
	force := res.Bool('f', "force")

	operands := paths.ExpandGlobs(res.Positional)

	// No operands: behave as a stdin-to-stdout filter, which is how these
	// tools are used inside pipelines.
	if len(operands) == 0 {
		data, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, err)
			return command.ExitFailure
		}
		if _, err := ctx.Stdout.Write(convertLineEndings(data, c.crlf)); err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, err)
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	// -n takes input/output pairs and never modifies the source.
	if res.Bool('n', "newfile") {
		if len(operands)%2 != 0 {
			fmt.Fprintf(ctx.Stderr, "%s: -n requires input and output file pairs\n", c.name)
			return command.ExitUsage
		}
		exit := command.ExitSuccess
		for i := 0; i < len(operands); i += 2 {
			if !c.convertToNewFile(ctx, operands[i], operands[i+1], quiet, force, keepDate) {
				exit = command.ExitFailure
			}
		}
		return exit
	}

	exit := command.ExitSuccess
	for _, name := range operands {
		if !c.convertInPlace(ctx, name, quiet, force, keepDate) {
			exit = command.ExitFailure
		}
	}
	return exit
}

func (c lineEndingCommand) convertInPlace(ctx *command.Context, name string, quiet, force, keepDate bool) bool {
	resolved, data, info, status := c.readSource(ctx, name, force, quiet)
	if status != sourceOK {
		return status == sourceSkipped
	}
	if !quiet {
		fmt.Fprintf(ctx.Stderr, "%s: converting file %s to %s format...\n", c.name, name, c.formatName())
	}

	converted := convertLineEndings(data, c.crlf)
	if bytes.Equal(converted, data) {
		// Nothing to do; leave the file (and its timestamp) untouched.
		return true
	}
	if err := replaceFile(resolved, converted, info); err != nil {
		output.Errorf(ctx.Stderr, c.name, "cannot write", name, err)
		return false
	}
	if keepDate {
		if err := os.Chtimes(resolved, info.ModTime(), info.ModTime()); err != nil {
			output.Errorf(ctx.Stderr, c.name, "cannot preserve timestamp of", name, err)
			return false
		}
	}
	return true
}

func (c lineEndingCommand) convertToNewFile(ctx *command.Context, in, out string, quiet, force, keepDate bool) bool {
	_, data, info, status := c.readSource(ctx, in, force, quiet)
	if status != sourceOK {
		return status == sourceSkipped
	}
	if !quiet {
		fmt.Fprintf(ctx.Stderr, "%s: converting file %s to file %s in %s format...\n", c.name, in, out, c.formatName())
	}

	target, err := paths.Resolve(out)
	if err != nil {
		output.Errorf(ctx.Stderr, c.name, "cannot write", out, err)
		return false
	}
	if err := os.WriteFile(target, convertLineEndings(data, c.crlf), info.Mode().Perm()); err != nil {
		output.Errorf(ctx.Stderr, c.name, "cannot write", out, err)
		return false
	}
	if keepDate {
		if err := os.Chtimes(target, info.ModTime(), info.ModTime()); err != nil {
			output.Errorf(ctx.Stderr, c.name, "cannot preserve timestamp of", out, err)
			return false
		}
	}
	return true
}

// Outcomes of loading a file for conversion. Skipping a binary file is
// deliberate and must not be reported as a failure, so it is kept
// distinct from a genuine read error.
const (
	sourceOK = iota
	sourceSkipped
	sourceError
)

// readSource loads a file and applies the binary-file guard.
func (c lineEndingCommand) readSource(ctx *command.Context, name string, force, quiet bool) (string, []byte, os.FileInfo, int) {
	resolved, err := paths.Resolve(name)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, c.name, name, err)
		return "", nil, nil, sourceError
	}
	info, err := os.Stat(resolved)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, c.name, name, err)
		return "", nil, nil, sourceError
	}
	if info.IsDir() {
		fmt.Fprintf(ctx.Stderr, "%s: %s: Is a directory\n", c.name, name)
		return "", nil, nil, sourceError
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		output.SimpleErrorf(ctx.Stderr, c.name, name, err)
		return "", nil, nil, sourceError
	}
	// Rewriting a binary file's CR bytes would corrupt it, so skip
	// anything containing a NUL unless the user insists with -f.
	if !force && bytes.IndexByte(data, 0) >= 0 {
		if !quiet {
			fmt.Fprintf(ctx.Stderr, "%s: Skipping binary file %s\n", c.name, name)
		}
		return "", data, info, sourceSkipped
	}
	return resolved, data, info, sourceOK
}

func (c lineEndingCommand) formatName() string {
	if c.crlf {
		return "DOS"
	}
	return "Unix"
}

// convertLineEndings rewrites every line terminator to LF, then to CRLF
// when that is the target. Normalizing first is what keeps unix2dos
// idempotent on files that already use CRLF.
func convertLineEndings(data []byte, toCRLF bool) []byte {
	unix := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !toCRLF {
		return unix
	}
	return bytes.ReplaceAll(unix, []byte("\n"), []byte("\r\n"))
}

// replaceFile writes data over path via a temporary file in the same
// directory, so an interrupted run cannot leave a half-converted file
// behind. The original permissions are carried over.
func replaceFile(path string, data []byte, info os.FileInfo) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".linuxcmd-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func init() {
	command.Register(lineEndingCommand{
		name:    "dos2unix",
		summary: "convert DOS line endings (CRLF) to Unix (LF)",
		crlf:    false,
	})
	command.Register(lineEndingCommand{
		name:    "unix2dos",
		summary: "convert Unix line endings (LF) to DOS (CRLF)",
		crlf:    true,
	})
}
