package main

import (
	"flag"
	"fmt"
	"os"
)

// kulang:
//   build  run build script to build the target

func main() {
	switch len(os.Args) {
	case 0:
		fmt.Printf("arg[0] must be command\n")
		os.Exit(KULANG_SUCCESS)
	case 1:
		os.Args = append(os.Args, "help")
	}

	subCmdName := os.Args[1]
	subCmd, ok := commands[subCmdName]
	if !ok {
		fmt.Printf("[Error] '%s' is not a recognized command\n", os.Args[1])
		os.Exit(KULANG_ERROR)
	}

	fs := subCmd.Flags
	if fs == nil {
		fs = flag.NewFlagSet(subCmd.Name, flag.ExitOnError)
	}

	err := fs.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err)
		os.Exit(KULANG_ERROR)
	}

	ret, err := subCmd.Func(Flags{fs})
	if err != nil {
		os.Exit(KULANG_ERROR)
	}

	os.Exit(ret)
}
