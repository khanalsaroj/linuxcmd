// Package command defines the shared abstractions every linuxcmd command
// implements, plus the registry that maps a command name to its
// implementation. cmd/linuxcmd is the only binary that gets built; every
// per-command .exe (ls.exe, cp.exe, ...) is a copy/hardlink of it that
// dispatches based on its own invoked filename.
package command

import (
	"io"
	"os"
	"sort"
)

// Standard process exit codes, mirroring common Unix shell conventions.
const (
	ExitSuccess    = 0
	ExitFailure    = 1
	ExitUsage      = 2
	ExitNotFound   = 127
	ExitInterrupt  = 130
)

// Context carries everything a command needs to run: its arguments and
// I/O streams. Streams are injected (rather than commands using os.Stdout
// directly) so commands stay testable and so pipelines can be wired up
// later without changing command code.
type Context struct {
	// CommandName is the name the command was invoked as (e.g. "ls"),
	// independent of the actual binary filename.
	CommandName string
	// Args are the raw arguments following the command name, exactly as
	// received from the OS (already tokenized by Windows' own
	// CreateProcess/argv splitting, so quoted arguments arrive pre-split).
	Args []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewContext builds a Context wired to the real OS streams.
func NewContext(name string, args []string) *Context {
	return &Context{
		CommandName: name,
		Args:        args,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

// Command is the interface every linuxcmd command implements.
type Command interface {
	// Name is the canonical Linux command name, e.g. "ls".
	Name() string
	// Summary is a one-line description used in help/usage listings.
	Summary() string
	// Run executes the command and returns a process exit code.
	Run(ctx *Context) int
}

var registry = map[string]Command{}

// Register adds a command to the global registry. It is called from the
// init() function of each file in the commands package. Registering under
// a name that already exists overwrites the previous entry, which lets
// tests substitute fakes if ever needed.
func Register(c Command) {
	registry[c.Name()] = c
}

// Lookup returns the command registered under name, if any.
func Lookup(name string) (Command, bool) {
	c, ok := registry[name]
	return c, ok
}

// Names returns every registered command name, sorted alphabetically.
func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
