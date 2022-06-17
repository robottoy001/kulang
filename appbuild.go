package main

import (
	"fmt"
	"log"
	"path"
)

type BuildOption struct {
	ConfigFile string
	BuildDir   string
	Targets    []string
}

type AppBuild struct {
	Option *BuildOption
	Scope  *Scope

	// all nodes map
	Nodes    map[string]*Node
	Defaults []*Node
	Pools    map[string]*Pool
	Edges    []*Edge
}

func NewAppBuild(option *BuildOption) *AppBuild {
	return &AppBuild{
		Option:   option,
		Scope:    NewScope(nil),
		Nodes:    make(map[string]*Node),
		Defaults: []*Node{},
		Pools:    make(map[string]*Pool),
		Edges:    []*Edge{},
	}
}

func (self *AppBuild) Initialize() {
	self.Scope.AppendRule(PhonyRule.Name, PhonyRule)
}

func (self *AppBuild) RunBuild() error {

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
		log.Fatal(err)
		return err
	}

	fmt.Println("start building...")
	err = self._RunBuild()
	if err != nil {
		return err
	}
	return nil
}

func (self *AppBuild) FindNode(path string) *Node {
	if node, ok := self.Nodes[path]; ok {
		return node
	}
	return NewNode(path)
}

func (self *AppBuild) AddDefaults(path string) {
	self.Defaults = append(self.Defaults, self.FindNode(path))
}

func (self *AppBuild) AddPool(poolName string, depth int) {
	self.Pools[poolName] = NewPool(poolName, depth)
}

func (self *AppBuild) FindPool(poolName string) *Pool {
	if pool, ok := self.Pools[poolName]; ok {
		return pool
	}
	return nil
}

func (self *AppBuild) AddBuild(edge *Edge) {
	self.Edges = append(self.Edges, edge)
}

func (self *AppBuild) AddOut(edge *Edge, path string) {
	node := self.FindNode(path)
	// TODO: ignore if exist
	node.InEdge = edge
	edge.Outs = append(edge.Outs, node)
}

func (self AppBuild) AddIn(edge *Edge, path string) {
	node := self.FindNode(path)
	node.OutEdges = append(node.OutEdges, edge)

	edge.Ins = append(edge.Ins, node)
}

func (self *AppBuild) getTargets() []*Node {
	var nodesToBuild []*Node
	if len(self.Option.Targets) > 0 {
		// return self.Option.Targets
		for _, path := range self.Option.Targets {
			if node := self.FindNode(path); node != nil {
				nodesToBuild = append(nodesToBuild, node)
			}
		}
		return nodesToBuild
	}

	if len(self.Defaults) > 0 {
		return self.Defaults
	}

	for _, e := range self.Edges {
		for _, o := range e.Outs {
			if len(o.OutEdges) == 0 {
				nodesToBuild = append(nodesToBuild, o)
			}
		}
	}

	return nodesToBuild
}

func (self *AppBuild) _RunBuild() error {
	// find target
	// 1. from commandline
	// 2. from default rule
	// 3. root node which don't have out edge
	targets := self.getTargets()
	for _, t := range targets {
		fmt.Printf("%s\n", t.Path)
	}

	return nil
}
