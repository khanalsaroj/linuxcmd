package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type curlCommand struct{}

func (curlCommand) Name() string    { return "curl" }
func (curlCommand) Summary() string { return "transfer data from a URL" }

var curlSpec = parser.Spec{
	{Short: 'o', HasArg: true},
	{Short: 'O'},
	{Short: 'L'},
	{Short: 'H', HasArg: true},
	{Short: 'X', HasArg: true},
	{Short: 's'},
	{Short: 'I'},
}

func (curlCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, curlSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "curl: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: curl [-o FILE] [-O] [-L] [-H HEADER] [-X METHOD] [-I] URL")
		return command.ExitUsage
	}
	url := res.Positional[0]

	method := "GET"
	if v, ok := res.Value('X', ""); ok {
		method = v
	}
	if res.Bool('I', "") {
		method = "HEAD"
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "curl: %s\n", err)
		return command.ExitFailure
	}
	if v, ok := res.Value('H', ""); ok {
		if idx := strings.IndexByte(v, ':'); idx >= 0 {
			req.Header.Set(strings.TrimSpace(v[:idx]), strings.TrimSpace(v[idx+1:]))
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	if !res.Bool('L', "") {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "curl: %s\n", err)
		return command.ExitFailure
	}
	defer resp.Body.Close()

	if res.Bool('I', "") {
		fmt.Fprintf(ctx.Stdout, "%s %s\n", resp.Proto, resp.Status)
		for k, vs := range resp.Header {
			for _, v := range vs {
				fmt.Fprintf(ctx.Stdout, "%s: %s\n", k, v)
			}
		}
		return command.ExitSuccess
	}

	var out io.Writer = ctx.Stdout
	outFile, hasOutFile := res.Value('o', "")
	if res.Bool('O', "") {
		outFile = filepath.Base(url)
		hasOutFile = true
	}
	if hasOutFile {
		resolved, err := paths.Resolve(outFile)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "curl: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		f, err := os.Create(resolved)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "curl: %s\n", output.LinuxErrorText(err))
			return command.ExitFailure
		}
		defer f.Close()
		out = f
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		fmt.Fprintf(ctx.Stderr, "curl: %s\n", err)
		return command.ExitFailure
	}
	if resp.StatusCode >= 400 {
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func init() { command.Register(curlCommand{}) }
