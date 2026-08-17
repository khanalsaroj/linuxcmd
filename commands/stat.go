package commands

import (
	"fmt"
	"os"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/paths"
)

type statCommand struct{}

func (statCommand) Name() string    { return "stat" }
func (statCommand) Summary() string { return "display file metadata" }

func (statCommand) Run(ctx *command.Context) int {
	if len(ctx.Args) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: stat FILE...")
		return command.ExitUsage
	}

	exit := command.ExitSuccess
	for _, arg := range paths.ExpandGlobs(ctx.Args) {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "stat", arg, err)
			exit = command.ExitFailure
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "stat", arg, err)
			exit = command.ExitFailure
			continue
		}

		fileType := "regular file"
		if info.IsDir() {
			fileType = "directory"
		}
		fmt.Fprintf(ctx.Stdout, "  File: %s\n", resolved)
		fmt.Fprintf(ctx.Stdout, "  Size: %-10d  Type: %s\n", info.Size(), fileType)
		fmt.Fprintf(ctx.Stdout, "Access: (%s)\n", output.FormatMode(info))
		fmt.Fprintf(ctx.Stdout, "Modify: %s\n", info.ModTime().Format("2006-01-02 15:04:05 -0700"))
	}
	return exit
}

func init() { command.Register(statCommand{}) }
