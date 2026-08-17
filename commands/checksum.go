package commands

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// checksumCommand implements the shared shape of sha256sum, sha1sum, and
// md5sum: hash each file (or stdin), print "<hex>  <name>" per GNU
// coreutils convention, and support -c to verify against such output.
type checksumCommand struct {
	name    string
	newHash func() hash.Hash
}

func (c checksumCommand) Name() string    { return c.name }
func (c checksumCommand) Summary() string { return "print or check " + c.name + " checksums" }

var checksumSpec = parser.Spec{
	{Short: 'c', Long: "check"},
}

func (c checksumCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, checksumSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, err)
		return command.ExitUsage
	}

	if res.Bool('c', "check") {
		return c.check(ctx, res.Positional)
	}

	files := paths.ExpandGlobs(res.Positional)
	if len(files) == 0 {
		sum, err := c.hashReader(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, output.LinuxErrorText(err))
			return command.ExitFailure
		}
		fmt.Fprintf(ctx.Stdout, "%s  -\n", sum)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, c.name, arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, c.name, arg, err)
			exit = command.ExitFailure
			continue
		}
		sum, err := c.hashReader(f)
		f.Close()
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, c.name, arg, err)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%s  %s\n", sum, arg)
	}
	return exit
}

func (c checksumCommand) hashReader(r io.Reader) (string, error) {
	h := c.newHash()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c checksumCommand) check(ctx *command.Context, files []string) int {
	target := "-"
	if len(files) > 0 {
		target = files[0]
	}
	var in io.Reader = ctx.Stdin
	if target != "-" {
		resolved, err := paths.Resolve(target)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, c.name, target, err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, c.name, target, err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "%s: %s\n", c.name, output.LinuxErrorText(err))
		return command.ExitFailure
	}

	exit := command.ExitSuccess
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "  ", 2)
		if len(fields) != 2 {
			fields = strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
		}
		want, name := fields[0], fields[1]
		resolved, err := paths.Resolve(name)
		if err != nil {
			fmt.Fprintf(ctx.Stdout, "%s: FAILED open or read\n", name)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			fmt.Fprintf(ctx.Stdout, "%s: FAILED open or read\n", name)
			exit = command.ExitFailure
			continue
		}
		got, err := c.hashReader(f)
		f.Close()
		if err != nil || !strings.EqualFold(got, want) {
			fmt.Fprintf(ctx.Stdout, "%s: FAILED\n", name)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%s: OK\n", name)
	}
	return exit
}

type cksumCommand struct{}

func (cksumCommand) Name() string    { return "cksum" }
func (cksumCommand) Summary() string { return "print CRC checksum and byte count" }

func (cksumCommand) Run(ctx *command.Context) int {
	files := paths.ExpandGlobs(ctx.Args)
	if len(files) == 0 {
		data, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "cksum: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		fmt.Fprintf(ctx.Stdout, "%d %d\n", crc32.ChecksumIEEE(data), len(data))
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cksum", arg, err)
			exit = command.ExitFailure
			continue
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "cksum", arg, err)
			exit = command.ExitFailure
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%d %d %s\n", crc32.ChecksumIEEE(data), len(data), arg)
	}
	return exit
}

func init() {
	command.Register(checksumCommand{name: "sha256sum", newHash: sha256.New})
	command.Register(checksumCommand{name: "sha1sum", newHash: sha1.New})
	command.Register(checksumCommand{name: "md5sum", newHash: md5.New})
	command.Register(cksumCommand{})
}
