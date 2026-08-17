package commands

import (
	"fmt"
	"os"
	"runtime"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type unameCommand struct{}

func (unameCommand) Name() string    { return "uname" }
func (unameCommand) Summary() string { return "print system information" }

var unameSpec = parser.Spec{
	{Short: 'a', Long: "all"},
	{Short: 's', Long: "kernel-name"},
	{Short: 'n', Long: "nodename"},
	{Short: 'r', Long: "kernel-release"},
	{Short: 'm', Long: "machine"},
}

func unameArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "i686"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}

func (unameCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, unameSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "uname: %s\n", err)
		return command.ExitUsage
	}

	host, _ := os.Hostname()
	kernelName := "Windows_NT"
	release := "10.0"
	machine := unameArch()

	all := res.Bool('a', "all")
	if all || (!res.Bool('s', "kernel-name") && !res.Bool('n', "nodename") && !res.Bool('r', "kernel-release") && !res.Bool('m', "machine")) {
		if all {
			fmt.Fprintf(ctx.Stdout, "%s %s %s %s\n", kernelName, host, release, machine)
		} else {
			fmt.Fprintln(ctx.Stdout, kernelName)
		}
		return command.ExitSuccess
	}

	var parts []string
	if res.Bool('s', "kernel-name") {
		parts = append(parts, kernelName)
	}
	if res.Bool('n', "nodename") {
		parts = append(parts, host)
	}
	if res.Bool('r', "kernel-release") {
		parts = append(parts, release)
	}
	if res.Bool('m', "machine") {
		parts = append(parts, machine)
	}
	for i, p := range parts {
		if i > 0 {
			fmt.Fprint(ctx.Stdout, " ")
		}
		fmt.Fprint(ctx.Stdout, p)
	}
	fmt.Fprintln(ctx.Stdout)
	return command.ExitSuccess
}

func init() { command.Register(unameCommand{}) }
