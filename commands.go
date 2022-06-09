package main

import "flag"

type Command struct {
	Name string
	Func CommandFunc

	Usage string
	Short string
	Long  string

	Flags *flag.FlagSet
}

type Flags struct {
	*flag.FlagSet
}

type CommandFunc func(Flags) (int, error)

var commands = make(map[string]Command)

func Commands() map[string]Command {
	return commands
}

func init() {
	RegisterCmd(Command{
		Name:  "help",
		Func:  cmdHelp,
		Usage: "<command>",
		Short: "Show help message",
	})
}

func RegisterCmd(cmd Command) {
	if cmd.Name == "" {
		panic("need name of command")
	}

	if cmd.Func == nil {
		panic("no command function")
	}

	if cmd.Short == "" {
		panic("need short help message")
	}

	if _, exist := commands[cmd.Name]; exist {
		panic("command has been registered")
	}

	commands[cmd.Name] = cmd
}
