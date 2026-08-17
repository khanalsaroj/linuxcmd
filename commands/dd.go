package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type ddCommand struct{}

func (ddCommand) Name() string    { return "dd" }
func (ddCommand) Summary() string { return "copy fixed-size blocks between files" }

// parseDDSize parses dd's bs=/count=-style sizes, which allow a trailing
// K/M/G unit (base 1024, unlike bs=1MB-style SI units).
func parseDDSize(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	multiplier := int64(1)
	last := s[len(s)-1]
	switch last {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid size '%s'", s)
	}
	return n * multiplier, nil
}

func (ddCommand) Run(ctx *command.Context) int {
	var ifPath, ofPath string
	blockSize := int64(512)
	skip, seek, count := int64(0), int64(0), int64(-1)

	for _, a := range ctx.Args {
		eq := strings.IndexByte(a, '=')
		if eq < 0 {
			fmt.Fprintf(ctx.Stderr, "dd: unrecognized operand '%s'\n", a)
			return command.ExitUsage
		}
		key, val := a[:eq], a[eq+1:]
		var err error
		switch key {
		case "if":
			ifPath = val
		case "of":
			ofPath = val
		case "bs":
			blockSize, err = parseDDSize(val)
		case "skip":
			skip, err = parseDDSize(val)
		case "seek":
			seek, err = parseDDSize(val)
		case "count":
			count, err = parseDDSize(val)
		default:
			fmt.Fprintf(ctx.Stderr, "dd: unrecognized operand '%s'\n", a)
			return command.ExitUsage
		}
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", err)
			return command.ExitUsage
		}
	}
	if blockSize == 0 {
		fmt.Fprintln(ctx.Stderr, "dd: bs must be greater than zero")
		return command.ExitUsage
	}

	var in io.Reader = ctx.Stdin
	if ifPath != "" {
		resolved, err := paths.Resolve(ifPath)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		defer f.Close()
		if skip > 0 {
			if _, err := f.Seek(skip*blockSize, io.SeekStart); err != nil {
				fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
				return command.ExitFailure
			}
		}
		in = f
	}

	var out io.Writer = ctx.Stdout
	if ofPath != "" {
		resolved, err := paths.Resolve(ofPath)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		flags := os.O_WRONLY | os.O_CREATE
		if seek == 0 {
			flags |= os.O_TRUNC
		}
		f, err := os.OpenFile(resolved, flags, 0644)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		defer f.Close()
		if seek > 0 {
			if _, err := f.Seek(seek*blockSize, io.SeekStart); err != nil {
				fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
				return command.ExitFailure
			}
		}
		out = f
	}

	buf := make([]byte, blockSize)
	var full, partial int64
	for count < 0 || full+partial < count {
		n, err := io.ReadFull(in, buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(werr))
				return command.ExitFailure
			}
			if int64(n) == blockSize {
				full++
			} else {
				partial++
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "dd: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
	}

	fmt.Fprintf(ctx.Stderr, "%d+%d records in\n%d+%d records out\n", full, partial, full, partial)
	return command.ExitSuccess
}

func init() { command.Register(ddCommand{}) }
