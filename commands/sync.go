package commands

import "linuxcmd/internal/command"

// syncCommand is a no-op stand-in for Unix "sync". Because linuxcmd runs
// each command as its own short-lived process, there is no persistent
// file descriptor state left over from other commands to flush, and
// Windows offers no simple equivalent of flushing every mounted volume's
// write cache without per-handle FlushFileBuffers calls. Documented as a
// no-op limitation in the README rather than silently pretending to sync
// disks it can't reach.
type syncCommand struct{}

func (syncCommand) Name() string    { return "sync" }
func (syncCommand) Summary() string { return "flush file system buffers (no-op on this build)" }

func (syncCommand) Run(*command.Context) int { return command.ExitSuccess }

func init() { command.Register(syncCommand{}) }
