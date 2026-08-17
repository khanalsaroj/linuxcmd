package commands

import (
	"fmt"

	"linuxcmd/internal/command"
)

// xzCommand is registered but not functional: Go's standard library has
// no XZ (LZMA2) encoder or decoder at all, unlike gzip/bzip2, so this
// would need an external dependency this project doesn't currently take
// on. Reporting that clearly beats silently doing nothing.
type xzCommand struct{}

func (xzCommand) Name() string    { return "xz" }
func (xzCommand) Summary() string { return "XZ compression (not supported in this build)" }

func (xzCommand) Run(ctx *command.Context) int {
	fmt.Fprintln(ctx.Stderr, "xz: not supported in this build (no XZ codec in Go's standard library; "+
		"would need an external dependency)")
	return command.ExitFailure
}

func init() { command.Register(xzCommand{}) }
