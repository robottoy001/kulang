package main

import (
	"fmt"
)

func cmdHelp(option *BuildOption, flags Flags) (int, error) {
	args := flags.Args()
	if len(args) == 0 {
		usage := `kulang is yet another Ninja build system
usage:
  kulang <command> [args...]
commands
`
		for _, cmd := range commands {
			usage += fmt.Sprintf("  %-15s %s\n", cmd.Name, cmd.Short)
		}
		fmt.Printf(usage)
		return KULANG_SUCCESS, nil
	} else if len(args) > 1 {
		return KULANG_ERROR, fmt.Errorf("only can be one command")
	}

	subCmd, ok := commands[args[0]]
	if !ok {
		return KULANG_ERROR, fmt.Errorf("no such command %s", args[0])
	}

	result := fmt.Sprintf("%s\n\nUsage:\n  kulang [option] %s   %s",
		subCmd.Short,
		subCmd.Name,
		subCmd.Usage)

	fmt.Printf("%s\n", result)

	if flagsText := dumpFlags(subCmd.Flags); flagsText != "" {
		fmt.Printf("flags:\n%s\n", flagsText)
	}

	return KULANG_SUCCESS, nil
}

func cmdBuild(option *BuildOption, flags Flags) (int, error) {
	fmt.Printf("%v\n", flags.Args())

	option.Targets = flags.Args()

	App := NewAppBuild(option)
	App.Initialize()
	err := App.RunBuild()
	if err != nil {
		return KULANG_ERROR, err
	}
	return KULANG_SUCCESS, nil
}

func cmdVersion(option *BuildOption, flags Flags) (int, error) {
	const (
		version = "0.0.1"
	)
	fmt.Printf("kulang %s\n", version)
	return KULANG_SUCCESS, nil
}

func cmdTargets(option *BuildOption, flags Flags) (int, error) {

	option.Targets = flags.Args()

	App := NewAppBuild(option)
	App.Initialize()
	err := App.Targets()
	if err != nil {
		return KULANG_ERROR, err
	}
	return KULANG_SUCCESS, nil
}

func cmdClean(option *BuildOption, flags Flags) (int, error) {
	option.Targets = flags.Args()

	App := NewAppBuild(option)
	App.Initialize()
	err := App.Clean()

	if err != nil {
		return KULANG_ERROR, err
	}
	return KULANG_SUCCESS, nil

	return KULANG_SUCCESS, nil
}
