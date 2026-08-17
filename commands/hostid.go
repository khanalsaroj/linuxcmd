package commands

import (
	"fmt"
	"hash/fnv"
	"os"

	"linuxcmd/internal/command"
)

type hostidCommand struct{}

func (hostidCommand) Name() string    { return "hostid" }
func (hostidCommand) Summary() string { return "print a machine-derived identifier" }

func (hostidCommand) Run(ctx *command.Context) int {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	h := fnv.New32a()
	h.Write([]byte(host))
	// Not Linux's actual hostid algorithm (derived from the resolved IP
	// address); this is a stable, host-derived stand-in documented here
	// and in the README.
	fmt.Fprintf(ctx.Stdout, "%08x\n", h.Sum32())
	return command.ExitSuccess
}

func init() { command.Register(hostidCommand{}) }
