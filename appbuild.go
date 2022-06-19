package main

import (
	"fmt"
	"log"
	"os"
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
	err := os.MkdirAll(self.Option.BuildDir, os.ModePerm)
	if err != nil {
		log.Fatal("Build directory error: ", err)
		return
	}
	if err != os.Chdir(self.Option.BuildDir) {
		log.Fatal("Change directory to ", self.Option.BuildDir, " faild")
	}
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
	n := NewNode(path)
	self.Nodes[path] = n
	return n
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

func CollectOutPutDitryNodes(edge *Edge, most_recent_input *Node) bool {
	for _, o := range edge.Outs {
		if edge.Rule.Name == "phony" {
			if edge.Ins == nil && !o.Exist() {
				return true
			}
			// TODO: xx update phony mtime
			return false
		}
		if !o.Exist() {
			return true
		}

		if most_recent_input != nil && most_recent_input.Status.MTime.After(o.Status.MTime) {
			// TODO: xx restart property
			fmt.Printf("xxxxxxx - Path： %s, most recent: %v\n", o.Path, most_recent_input.Status.MTime)
			return true
		}
	}
	return false
}

func (self *AppBuild) CollectDitryNodes(node *Node) bool {
	// leaf node
	if node.InEdge == nil {
		if node.StatusKnow() {
			return true
		}
		if ok := node.Stat(); !ok {
			return false
		}
		if exist := node.Exist(); !exist {
			fmt.Printf("%s has build line, but missing", node.Path)
		}
		// mark dirty if no exist
		node.SetDirty(!node.Exist())
		return true
	}

	// update output mode time & exist status
	for _, o := range node.InEdge.Outs {
		if ok := o.Stat(); !ok {
			fmt.Printf("out stat err(%v) %s\n", ok, o.Path)
			return false
		}
	}

	// if any input is dirty, current node is dirty
	var most_recent_input *Node = nil
	var dirty bool = false
	for _, i := range node.InEdge.Ins {
		if ok := self.CollectDitryNodes(i); !ok {
			return false
		}
		if i.InEdge != nil && !i.InEdge.OutPutReady {
			node.InEdge.OutPutReady = false
		}

		// TODO:xxx consider order only
		if i.Status.Dirty {
			dirty = true
		} else {
			if most_recent_input == nil || most_recent_input.Status.MTime.After(i.Status.MTime) {
				most_recent_input = i
			}
		}
	}
	//fmt.Printf("Most:recent:mtime, %v - path:%s, mode time:%v - node: %s\n",
	//	most_recent_input.Status.MTime, most_recent_input.Path, i.Status.MTime, i.Path)
	if most_recent_input != nil {

		//fmt.Printf("Curnode: %s Most:recent:mtime, %v - path:%s\n", node.Path,
		//	most_recent_input.Status.MTime, most_recent_input.Path)
	}

	if !dirty {
		dirty = CollectOutPutDitryNodes(node.InEdge, most_recent_input)
	}

	// make dirty
	for _, o := range node.InEdge.Outs {
		o.SetDirty(dirty)
		if dirty {
			fmt.Printf("dirty(%v) %s\n", dirty, node.Path)
		}
	}

	return true
}

func (self *AppBuild) _RunBuild() error {
	// find target
	// 1. from commandline
	// 2. from default rule
	// 3. root node which don't have out edge
	targets := self.getTargets()
	var allDitryNodes []*Node
	for _, t := range targets {
		self.CollectDitryNodes(t)
	}

	for _, node := range allDitryNodes {
		if node.Stat() {
			fmt.Printf("exist: %v ", node.Status.MTime)
		}
		fmt.Printf("%d - %s\n", node.Status.Exist, node.Path)
	}

	return nil
}
