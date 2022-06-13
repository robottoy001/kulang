package main

import (
	"fmt"
	"log"
	"path"
)

type BuildOption struct {
	ConfigFile string
	BuildDir   string
}

type AppBuild struct {
	Option *BuildOption
	Scope  *Scope
}

func NewAppBuild(option *BuildOption) *AppBuild {
	return &AppBuild{
		Option: option,
		Scope: &Scope{
			Rules:  make(map[string]*Rule),
			Vars:   make(map[string]string),
			Parent: nil,
		},
	}
}

func (self *AppBuild) RunBuild() error {
	fmt.Println("start building...")

	// parser load .ninja file
	// default, Ninja parser & scanner
	p := NewParser(self.Scope)
	err := p.Load(path.Join(self.Option.BuildDir, self.Option.ConfigFile))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = p.Parse()
	if err != nil {
		return err
	}

	err = self._RunBuild()
	if err != nil {
		return err
	}
	return nil
}

func (self *AppBuild) _RunBuild() error {
	return nil
}
