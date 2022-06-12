package main

import (
	"bytes"
	"flag"
)

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

func dumpFlags(fs *flag.FlagSet) string {
	if fs == nil {
		return ""
	}

	out := fs.Output()
	defer fs.SetOutput(out)

	buf := new(bytes.Buffer)
	fs.SetOutput(buf)
	fs.PrintDefaults()

	return buf.String()
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
	RegisterCmd(Command{
		Name:  "build",
		Func:  cmdBuild,
		Usage: "[-D <DIR>] [--config=<ninja file>] [targets...]",
		Short: "Build targets which specified",
		Flags: func() *flag.FlagSet {
			fs := flag.NewFlagSet("build", flag.ExitOnError)
			fs.String("D", ".", "directory which include .ninja file")
			fs.String("config", "build.ninja", "specified .ninja file")
			return fs
		}(),
	})
	RegisterCmd(Command{
		Name:  "version",
		Func:  cmdVersion,
		Usage: "",
		Short: "Show version of kulang",
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
