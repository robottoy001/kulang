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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

func enableCPUProfile(cpuProfilePath string) (closer func()) {
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
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)

	stop := enableCPUProfile("./cpu.profile")
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT:
				fmt.Printf("Signal: %v, quit.\n", s)
				stop()
				os.Exit(KulangError)
			default:
				fmt.Printf("Got other signal, %v\n", s)
			}
		}
	}()

	switch len(os.Args) {
	case 0:
		fmt.Printf("arg[0] must be command\n")
		os.Exit(KulangSuccess)
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
		os.Exit(KulangError)
	}

	fs := subCmd.Flags
	if fs == nil {
		fs = flag.NewFlagSet(subCmd.Name, flag.ExitOnError)
	}

	err := fs.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(KulangError)
	}

	ret, err := subCmd.Func(option, Flags{fs})
	if err != nil {
		os.Exit(KulangError)
	}

	stop()

	os.Exit(ret)
}
