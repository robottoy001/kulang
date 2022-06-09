package main

import "fmt"

func cmdHelp(flags Flags) (int, error) {
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
	}

	return KULANG_SUCCESS, nil
}
