package commands

import "linuxcmd/internal/command"

type trueCommand struct{}

func (trueCommand) Name() string             { return "true" }
func (trueCommand) Summary() string          { return "always succeed" }
func (trueCommand) Run(*command.Context) int { return command.ExitSuccess }

type falseCommand struct{}

func (falseCommand) Name() string             { return "false" }
func (falseCommand) Summary() string          { return "always fail" }
func (falseCommand) Run(*command.Context) int { return command.ExitFailure }

func init() {
	command.Register(trueCommand{})
	command.Register(falseCommand{})
}
