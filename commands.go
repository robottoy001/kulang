/*
 * Copyright (c) 2022 Huawei Device Co., Ltd.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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

type CommandFunc func(*BuildOption, Flags) (int, error)

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
		Usage: "[targets...]",
		Short: "Build targets which specified",
	})

	RegisterCmd(Command{
		Name:  "targets",
		Func:  cmdTargets,
		Usage: "[-D <DIR>] [--config=<ninja file>] [switch...]",
		Short: "Show the targets need to build",
		Flags: func() *flag.FlagSet {
			fs := flag.NewFlagSet("build", flag.ExitOnError)
			fs.Bool("rule", false, "print all rules")
			return fs
		}(),
	})
	RegisterCmd(Command{
		Name:  "clean",
		Func:  cmdClean,
		Usage: "Clean built files",
		Short: "Clean built files",
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
