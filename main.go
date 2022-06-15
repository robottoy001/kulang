package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func enableCpuProfile(cpuProfilePath string) (closer func()) {
	closer = func() {}
	if cpuProfilePath != "" {
		f, err := os.Create(cpuProfilePath)
		if err != nil {
			log.Panicf("could not create cpu profile: %v", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Panicf("error: %v", err)
		}
		closer = pprof.StopCPUProfile
	}
	runtime.SetBlockProfileRate(20)
	return
}

// kulang:
//   build  run build script to build the target

func main() {
	stop := enableCpuProfile("./cpu.profile")

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

	stop()

	os.Exit(ret)
}
