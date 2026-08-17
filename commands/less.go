package commands

import "linuxcmd/internal/command"

// lessCommand is currently a "more"-equivalent pager: it pages output on
// a real console and streams straight through when redirected, but does
// not yet implement less' backward scrolling or "/search". Documented as
// a subset in the README; full raw-mode terminal handling (arrow keys,
// search, "q" to quit mid-file) is a larger follow-up.
type lessCommand struct{}

func (lessCommand) Name() string    { return "less" }
func (lessCommand) Summary() string { return "page through file contents (more-equivalent subset)" }

func (lessCommand) Run(ctx *command.Context) int {
	return moreCommand{}.Run(ctx)
}

func init() { command.Register(lessCommand{}) }
