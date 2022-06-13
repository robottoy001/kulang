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

	// all nodes map
	Nodes    map[string]*Node
	Defaults []*Node
}

func NewAppBuild(option *BuildOption) *AppBuild {
	return &AppBuild{
		Option: option,
		Scope: &Scope{
			Rules:  make(map[string]*Rule),
			Vars:   make(map[string]string),
			Parent: nil,
		},
		Nodes:    make(map[string]*Node),
		Defaults: []*Node{},
	}
}

func (self *AppBuild) RunBuild() error {
	fmt.Println("start building...")

	// parser load .ninja file
	// default, Ninja parser & scanner
	p := NewParser(self)
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

func (self *AppBuild) findNode(path string) *Node {
	if node, ok := self.Nodes[path]; ok {
		return node
	}
	return NewNode(path)
}

func (self *AppBuild) AddDefaults(path string) {
	self.Defaults = append(self.Defaults, self.findNode(path))
}

func (self *AppBuild) _RunBuild() error {
	return nil
}
