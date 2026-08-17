package commands

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// opensslCommand implements only a small, safe subset -- "rand" and
// "dgst" -- rather than a full OpenSSL CLI or delegating to a real
// openssl.exe that may not be installed.
type opensslCommand struct{}

func (opensslCommand) Name() string    { return "openssl" }
func (opensslCommand) Summary() string { return "rand/dgst subcommands (small safe subset of OpenSSL)" }

func (opensslCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: openssl {rand -hex N | dgst -sha256|-sha1|-md5 [FILE]}")
		return command.ExitUsage
	}
	switch ctx.Args[0] {
	case "rand":
		return opensslRand(ctx, ctx.Args[1:])
	case "dgst":
		return opensslDgst(ctx, ctx.Args[1:])
	default:
		fmt.Fprintf(ctx.Stderr, "openssl: unsupported subcommand '%s'\n", ctx.Args[0])
		return command.ExitUsage
	}
}

func opensslRand(ctx *command.Context, args []string) int {
	hexMode := false
	var n int
	found := false
	for i := 0; i < len(args); i++ {
		if args[i] == "-hex" {
			hexMode = true
			continue
		}
		v, err := strconv.Atoi(args[i])
		if err == nil {
			n = v
			found = true
		}
	}
	if !found || n < 1 {
		fmt.Fprintln(ctx.Stderr, "usage: openssl rand -hex N")
		return command.ExitUsage
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		fmt.Fprintf(ctx.Stderr, "openssl: %s\n", err)
		return command.ExitFailure
	}
	if hexMode {
		fmt.Fprintln(ctx.Stdout, hex.EncodeToString(buf))
	} else {
		ctx.Stdout.Write(buf)
	}
	return command.ExitSuccess
}

func opensslDgst(ctx *command.Context, args []string) int {
	var h hash.Hash
	var file string
	for _, a := range args {
		switch a {
		case "-sha256":
			h = sha256.New()
		case "-sha1":
			h = sha1.New()
		case "-md5":
			h = md5.New()
		default:
			file = a
		}
	}
	if h == nil {
		h = sha256.New()
	}

	var in io.Reader = ctx.Stdin
	if file != "" {
		resolved, err := paths.Resolve(file)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "openssl", file, err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "openssl", file, err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	if _, err := io.Copy(h, in); err != nil {
		fmt.Fprintf(ctx.Stderr, "openssl: %s\n", err)
		return command.ExitFailure
	}
	if file != "" {
		fmt.Fprintf(ctx.Stdout, "%x  %s\n", h.Sum(nil), file)
	} else {
		fmt.Fprintf(ctx.Stdout, "%x\n", h.Sum(nil))
	}
	return command.ExitSuccess
}

func init() { command.Register(opensslCommand{}) }
