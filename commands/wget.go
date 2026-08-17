package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

type wgetCommand struct{}

func (wgetCommand) Name() string    { return "wget" }
func (wgetCommand) Summary() string { return "download a URL to a file" }

var wgetSpec = parser.Spec{
	{Short: 'O', HasArg: true},
}

func (wgetCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, wgetSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wget: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: wget [-O FILE] URL")
		return command.ExitUsage
	}
	url := res.Positional[0]

	outName, ok := res.Value('O', "")
	if !ok {
		outName = filepath.Base(url)
		if outName == "" || outName == "." || outName == "/" {
			outName = "index.html"
		}
	}
	resolved, err := paths.Resolve(outName)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wget: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wget: %s\n", err)
		return command.ExitFailure
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Fprintf(ctx.Stderr, "wget: server returned %s\n", resp.Status)
		return command.ExitFailure
	}

	f, err := os.Create(resolved)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wget: %s\n", output.LinuxErrorText(err))
		return command.ExitFailure
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "wget: %s\n", err)
		return command.ExitFailure
	}
	fmt.Fprintf(ctx.Stderr, "saved '%s' (%d bytes)\n", outName, n)
	return command.ExitSuccess
}

func init() { command.Register(wgetCommand{}) }
