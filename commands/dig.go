package commands

import (
	"fmt"
	"net"
	"strings"

	"linuxcmd/internal/command"
)

type digCommand struct{}

func (digCommand) Name() string    { return "dig" }
func (digCommand) Summary() string { return "query DNS records for a name" }

func (digCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: dig NAME [TYPE]")
		return command.ExitUsage
	}
	name := ctx.Args[0]
	recordType := "A"
	if len(ctx.Args) >= 2 {
		recordType = strings.ToUpper(ctx.Args[1])
	}

	fmt.Fprintf(ctx.Stdout, ";; QUESTION SECTION:\n;%s.\t\tIN\t%s\n\n;; ANSWER SECTION:\n", name, recordType)

	switch recordType {
	case "A", "AAAA":
		addrs, err := net.LookupHost(name)
		if err != nil {
			fmt.Fprintln(ctx.Stderr, "dig: query failed")
			return command.ExitFailure
		}
		for _, a := range addrs {
			isV4 := net.ParseIP(a).To4() != nil
			if (recordType == "A") == isV4 {
				fmt.Fprintf(ctx.Stdout, "%s.\t\tIN\t%s\t%s\n", name, recordType, a)
			}
		}
	case "MX":
		records, err := net.LookupMX(name)
		if err != nil {
			fmt.Fprintln(ctx.Stderr, "dig: query failed")
			return command.ExitFailure
		}
		for _, r := range records {
			fmt.Fprintf(ctx.Stdout, "%s.\t\tIN\tMX\t%d %s\n", name, r.Pref, r.Host)
		}
	case "TXT":
		records, err := net.LookupTXT(name)
		if err != nil {
			fmt.Fprintln(ctx.Stderr, "dig: query failed")
			return command.ExitFailure
		}
		for _, r := range records {
			fmt.Fprintf(ctx.Stdout, "%s.\t\tIN\tTXT\t%q\n", name, r)
		}
	case "NS":
		records, err := net.LookupNS(name)
		if err != nil {
			fmt.Fprintln(ctx.Stderr, "dig: query failed")
			return command.ExitFailure
		}
		for _, r := range records {
			fmt.Fprintf(ctx.Stdout, "%s.\t\tIN\tNS\t%s\n", name, r.Host)
		}
	default:
		fmt.Fprintf(ctx.Stderr, "dig: unsupported record type '%s'\n", recordType)
		return command.ExitUsage
	}
	return command.ExitSuccess
}

func init() { command.Register(digCommand{}) }
