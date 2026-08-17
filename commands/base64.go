package commands

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type base64Command struct{}

func (base64Command) Name() string    { return "base64" }
func (base64Command) Summary() string { return "encode or decode base64" }

var base64Spec = parser.Spec{
	{Short: 'd', Long: "decode"},
}

func (base64Command) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, base64Spec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "base64: %s\n", err)
		return command.ExitUsage
	}
	decode := res.Bool('d', "decode")

	var in io.Reader = ctx.Stdin
	files := paths.ExpandGlobs(res.Positional)
	if len(files) > 0 && files[0] != "-" {
		resolved, err := paths.Resolve(files[0])
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "base64", files[0], err)
			return command.ExitFailure
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "base64", files[0], err)
			return command.ExitFailure
		}
		defer f.Close()
		in = f
	}

	data, err := io.ReadAll(in)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "base64: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	if decode {
		decoded, err := base64.StdEncoding.DecodeString(trimBase64Whitespace(string(data)))
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "base64: invalid input\n")
			return command.ExitFailure
		}
		ctx.Stdout.Write(decoded)
		return command.ExitSuccess
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		fmt.Fprintln(ctx.Stdout, encoded[i:end])
	}
	return command.ExitSuccess
}

func trimBase64Whitespace(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' || c == '\r' || c == ' ' || c == '\t' {
			continue
		}
		out = append(out, c)
	}
	return string(out)
}

func init() { command.Register(base64Command{}) }
