package commands

import (
	"bufio"
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

// moreCommand is a simple page-at-a-time viewer. Real "more" pauses for
// a keypress between pages on an interactive console; here it prints a
// "-- More (N%) --" marker between pages when stdout is a real console,
// and streams straight through (no paging) otherwise, matching how ls
// already distinguishes interactive vs redirected output.
type moreCommand struct{}

func (moreCommand) Name() string    { return "more" }
func (moreCommand) Summary() string { return "page through file contents" }

const morePageLines = 24

func (moreCommand) Run(ctx *command.Context) int {
	process := func(r *bufio.Scanner) {
		interactive := isInteractive(ctx)
		count := 0
		for r.Scan() {
			fmt.Fprintln(ctx.Stdout, r.Text())
			count++
			if interactive && count%morePageLines == 0 {
				fmt.Fprint(ctx.Stdout, "-- More --")
				var discard [1]byte
				ctx.Stdin.Read(discard[:])
				fmt.Fprint(ctx.Stdout, "\r            \r")
			}
		}
	}

	files := paths.ExpandGlobs(ctx.Args)
	if len(files) == 0 {
		scanner := bufio.NewScanner(ctx.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "more", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "more", arg, err)
			exit = command.ExitFailure
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		f.Close()
	}
	return exit
}

func init() { command.Register(moreCommand{}) }
