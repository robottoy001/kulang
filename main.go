package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"
)

var workDirectory string
var configFile string

func init() {
	flag.StringVar(&workDirectory, "C", ".", "directory which include .ninja file")
	flag.StringVar(&configFile, "f", "build.ninja", "specified .ninja file")
}

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
	c := make(chan os.Signal)
	//signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	stop := enableCpuProfile("./cpu.profile")
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT:
				fmt.Printf("Signal: %v, quit.\n", s)
				stop()
				//os.Exit(KULANG_ERROR)
				panic("signaled")
			default:
				fmt.Printf("Got other signal, %v\n", s)
			}
		}
	}()

	switch len(os.Args) {
	case 0:
		fmt.Printf("arg[0] must be command\n")
		os.Exit(KULANG_SUCCESS)
	case 1:
		os.Args = append(os.Args, "help")
	}

	// main options
	flag.Parse()

	option := &BuildOption{
		BuildDir:   workDirectory,
		ConfigFile: configFile,
		Targets:    []string{},
	}

	//subCmdName := os.Args[1]
	args := flag.Args()
	if len(args) == 0 {
		args = append(args, "help")
	}
	subCmdName := args[0]
	subCmd, ok := commands[subCmdName]
	if !ok {
		fmt.Printf("[Error] '%s' is not a recognized command\n", os.Args[1])
		os.Exit(KULANG_ERROR)
	}

	fs := subCmd.Flags
	if fs == nil {
		fs = flag.NewFlagSet(subCmd.Name, flag.ExitOnError)
	}

	err := fs.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(KULANG_ERROR)
	}

	ret, err := subCmd.Func(option, Flags{fs})
	if err != nil {
		os.Exit(KULANG_ERROR)
	}

	stop()

	os.Exit(ret)
}
