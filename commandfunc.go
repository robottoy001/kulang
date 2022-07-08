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
	"fmt"
	"sort"

	"gitee.com/kulang/lib"
	"gitee.com/kulang/utils"
)

func cmdHelp(option *lib.BuildOption, flags Flags) (int, error) {
	args := flags.Args()
	if len(args) == 0 {
		usage := `kulang is yet another Ninja build system
usage:
  kulang [option] <command> [args...]
option:
`

		usage += dumpFlags(optionFlag)
		usage += "commands\n"

		var sortedCommands []string
		for _, cmd := range commands {
			sortedCommands = append(sortedCommands, cmd.Name)
		}

		sort.Sort(sort.StringSlice(sortedCommands))
		for _, cmdName := range sortedCommands {
			usage += fmt.Sprintf("  %-15s %s\n", cmdName, commands[cmdName].Short)
		}

		fmt.Printf(usage)
		return utils.KulangSuccess, nil
	} else if len(args) > 1 {
		return utils.KulangError, fmt.Errorf("only can be one command")
	}

	subCmd, ok := commands[args[0]]
	if !ok {
		return utils.KulangError, fmt.Errorf("no such command %s", args[0])
	}

	result := fmt.Sprintf("%s\n\nUsage:\n  kulang [option] %s %s",
		subCmd.Short,
		subCmd.Name,
		subCmd.Usage)

	fmt.Printf("%s\n", result)

	if flagsText := dumpFlags(subCmd.Flags); flagsText != "" {
		fmt.Printf("flags:\n%s\n", flagsText)
	}

	return utils.KulangSuccess, nil
}

func cmdBuild(option *lib.BuildOption, flags Flags) (int, error) {
	option.Targets = flags.Args()

	App := lib.NewAppBuild(option)
	App.Initialize()
	err := App.RunBuild()
	if err != nil {
		return utils.KulangError, err
	}
	return utils.KulangSuccess, nil
}

func cmdVersion(option *lib.BuildOption, flags Flags) (int, error) {
	const (
		version = "0.0.2"
	)
	fmt.Printf("kulang %s\n", version)
	return utils.KulangSuccess, nil
}

func cmdTargets(option *lib.BuildOption, flags Flags) (int, error) {

	option.Targets = flags.Args()

	App := lib.NewAppBuild(option)
	App.Initialize()
	err := App.Targets()
	if err != nil {
		return utils.KulangError, err
	}

	return utils.KulangSuccess, nil
}

func cmdClean(option *lib.BuildOption, flags Flags) (int, error) {
	option.Targets = flags.Args()

	App := lib.NewAppBuild(option)
	App.Initialize()
	err := App.Clean()

	if err != nil {
		return utils.KulangError, err
	}

	return utils.KulangSuccess, nil
}
