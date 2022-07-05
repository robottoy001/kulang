package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"syscall"
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
	Runner   *Runner
	Fs       FileSystem
}

func NewAppBuild(option *BuildOption) *AppBuild {
	return &AppBuild{
		Option:   option,
		Scope:    NewScope(nil),
		Nodes:    make(map[string]*Node),
		Defaults: []*Node{},
		Pools:    make(map[string]*Pool),
		Edges:    []*Edge{},
		Runner:   NewRunner(),
		Fs:       RealFileSystem{},
	}
}

func (b *AppBuild) Initialize() {
	//b.Scope.AppendRule(PhonyRule.Name, PhonyRule)
	absBuildDir, err := filepath.Abs(filepath.Dir(b.Option.BuildDir))
	if err != nil {
		log.Fatal("Get build directory absolute path fail:", err)
		return
	}
	err = os.MkdirAll(absBuildDir, os.ModePerm)
	if err != nil {
		log.Fatal("Build directory error: ", err)
		return
	}
	if err != os.Chdir(absBuildDir) {
		log.Fatal("Change directory to ", b.Option.BuildDir, " faild")
	}
}

func (b *AppBuild) RunBuild() error {

	// parser load .ninja file
	// default, Ninja parser & scanner
	p := NewParser(b, b.Scope)
	err := p.Load(path.Join(b.Option.BuildDir, b.Option.ConfigFile))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = p.Parse()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = b._RunBuild()
	if err != nil {
		return err
	}
	return nil
}

func (b *AppBuild) Targets() error {
	p := NewParser(b, b.Scope)
	err := p.Load(path.Join(b.Option.BuildDir, b.Option.ConfigFile))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = p.Parse()
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, e := range b.Edges {
		for _, o := range e.Outs {
			if len(o.OutEdges) == 0 && o.InEdge != nil {
				fmt.Printf("%s: %s\n", o.Path, o.InEdge.Rule.Name)
			}
		}
	}

	return nil
}

func (b *AppBuild) QueryNode(path string) *Node {
	if node, ok := b.Nodes[path]; ok {
		return node
	}
	return nil
}

func (b *AppBuild) FindNode(path string) *Node {
	if node, ok := b.Nodes[path]; ok {
		return node
	}
	n := NewNode(path)
	b.Nodes[path] = n
	return n
}

func (b *AppBuild) AddDefaults(path string) {
	b.Defaults = append(b.Defaults, b.FindNode(path))
}

func (b *AppBuild) AddPool(poolName string, depth int) {
	b.Pools[poolName] = NewPool(poolName, depth)
}

func (b *AppBuild) FindPool(poolName string) *Pool {
	if pool, ok := b.Pools[poolName]; ok {
		return pool
	}
	return nil
}

func (b *AppBuild) AddBuild(edge *Edge) {
	b.Edges = append(b.Edges, edge)
}

func (b *AppBuild) AddOut(edge *Edge, path string) {
	node := b.FindNode(path)
	// TODO: ignore if exist
	node.InEdge = edge
	edge.Outs = append(edge.Outs, node)
}

func (b AppBuild) AddIn(edge *Edge, path string) {
	node := b.FindNode(path)
	node.OutEdges = append(node.OutEdges, edge)

	edge.Ins = append(edge.Ins, node)
}

func (b AppBuild) AddValids(edge *Edge, path string) {
	node := b.FindNode(path)
	node.ValidOutEdges = append(node.ValidOutEdges, edge)

	edge.Validations = append(edge.Validations, node)
}

func (b *AppBuild) removeFiles(file string) error {
	err := os.Remove(file)
	if err == nil {
		return nil
	}

	if e, ok := err.(*os.PathError); ok {
		// doesn't exist, otherwise report & return error
		if e.Err == syscall.ENOENT {
			return err
		}
		fmt.Printf("Error: %s\n", e.Error())
	}
	return err
}

func (b *AppBuild) cleanAll() error {
	var cleaned = map[string]bool{}
	var proceeded int = 0
	for _, e := range b.Edges {
		if e.IsPhony() {
			continue
		}
		if e.QueryVar("generator") != "" {
			fmt.Printf("generator edge, %s\n", e.String())
			continue
		}
		for _, o := range e.Outs {
			if _, ok := cleaned[o.Path]; ok {
				continue
			}
			cleaned[o.Path] = true
			if b.removeFiles(o.Path) == nil {
				proceeded += 1
			}
		}
		// remove rsp file
		rspfile := e.QueryVar("rspfile")
		if rspfile != "" {
			if _, ok := cleaned[rspfile]; ok {
				continue
			}
			cleaned[rspfile] = true
			if b.removeFiles(rspfile) == nil {
				proceeded += 1
			}
		}
		// remve depfile
		depfile := e.QueryVar("depfile")
		if depfile != "" {
			if _, ok := cleaned[rspfile]; ok {
				continue
			}
			cleaned[depfile] = true
			if b.removeFiles(depfile) == nil {
				proceeded += 1
			}
		}

	}
	fmt.Printf("Cleaned %d files, %d edges\n", proceeded, len(b.Edges))
	return nil

}

func (b *AppBuild) Clean() error {

	p := NewParser(b, b.Scope)
	err := p.Load(path.Join(b.Option.BuildDir, b.Option.ConfigFile))
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = p.Parse()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// clean all
	if len(b.Option.Targets) == 0 {
		return b.cleanAll()
	}

	return nil
}

func (b *AppBuild) getTargets() []*Node {
	var nodesToBuild []*Node
	if len(b.Option.Targets) > 0 {
		// return b.Option.Targets
		for _, path := range b.Option.Targets {
			if node := b.QueryNode(path); node != nil {
				nodesToBuild = append(nodesToBuild, node)
			}
		}
		return nodesToBuild
	}

	if len(b.Defaults) > 0 {
		return b.Defaults
	}

	for _, e := range b.Edges {
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
		if edge.IsPhony() {
			if edge.Ins == nil && !o.Exist() {
				return true
			}
			// TODO: xx update phony mtime
			if most_recent_input != nil {
				o.SetPhonyMtime(most_recent_input.Status.MTime)
			}
			return false
		}
		if !o.Exist() {
			return true
		}

		if most_recent_input != nil && most_recent_input.Status.MTime.After(o.Status.MTime) {
			// TODO: xx restart property
			//fmt.Printf("xxxxxxx - Path： %s, most recent: %s, %v\n", o.Path, most_recent_input.Path, most_recent_input.Status.MTime)
			return true
		}
	}
	return false
}

func (b *AppBuild) verifyDAG(node *Node, stack []*Node) bool {
	if node.InEdge.VisitStatus != VISITED_IN_STACK {
		return true
	}

	var foundIndex int = -1
	for d := 0; d < len(stack); d += 1 {
		if stack[d].InEdge == node.InEdge {
			foundIndex = d
			break
		}
	}

	stack[foundIndex] = node

	var err string = "Dependency cycle: "
	for d := foundIndex; d < len(stack); d += 1 {
		err += stack[d].Path
		err += "->"
	}
	err += stack[foundIndex].Path

	log.Fatal(err)

	return false
}

func identPrint(node *Node, depth int, format string, a ...interface{}) {
	for i := 0; i < depth; i += 1 {
		fmt.Printf(" ")
	}
	fmt.Printf(format, a...)
}

// TODO:xxx need to visit validation nodes
// STATUS: visit if not
func (b *AppBuild) CollectDitryNodes(node *Node, stack []*Node) bool {
	//identPrint(node, len(stack), "Begin: %s\n", node.Path)
	// leaf node
	if node.InEdge == nil {
		if node.StatusKnow() {
			return true
		}

		// check if leaf node exist
		if ok := node.Stat(b.Fs); !ok {
			return false
		}
		if exist := node.Exist(); !exist {
			fmt.Printf("%s has build line, but missing\n", node.Path)
		}
		// mark dirty if no exist
		node.SetDirty(!node.Exist())
		return true
	}

	if node.InEdge.VisitStatus == VISITED_DONE {
		return true
	}

	if !b.verifyDAG(node, stack) {
		return false
	}

	// ninja trace visiting status
	node.InEdge.VisitStatus = VISITED_IN_STACK
	stack = append(stack, node)
	// initial OutPutReady true
	node.InEdge.OutPutReady = true
	// initial dirty flag
	var dirty bool = false

	// update output mode time & exist status
	for _, o := range node.InEdge.Outs {
		//fmt.Printf("    Outs: %s\n", o.Path)
		if ok := o.Stat(b.Fs); !ok {
			fmt.Printf("out stat err(%v) %s\n", ok, o.Path)
			return false
		}
	}

	// if any input is dirty, current node is dirty
	var most_recent_input *Node = nil
	for index, i := range node.InEdge.Ins {
		//fmt.Printf("    ints: %s\n", i.Path)
		if ok := b.CollectDitryNodes(i, stack); !ok {
			return false
		}

		if i.InEdge != nil && !i.InEdge.OutPutReady {
			node.InEdge.OutPutReady = false
		}

		if !node.InEdge.IsOrderOnly(index) {
			if i.Status.Dirty {
				dirty = true
			} else {
				if most_recent_input == nil || most_recent_input.Status.MTime.After(i.Status.MTime) {
					most_recent_input = i
				}
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
		//if dirty {
		//	fmt.Printf("set dirty %v: %s\n", o.Status.Dirty, o.Path)
		//	fmt.Printf("dirty(%v) cur node: %s\n", dirty, node.Path)
		//}
	}

	if dirty && !(node.InEdge.IsPhony() && len(node.InEdge.Ins) == 0) {
		//fmt.Printf("Node %s inEdge's output is false\n", node.Path)
		node.InEdge.OutPutReady = false
	}

	node.InEdge.VisitStatus = VISITED_DONE
	if stack[len(stack)-1] != node {
		log.Fatal("stack is bad")
	}
	stack = stack[:len(stack)-1]

	//identPrint(node, len(stack), "  end: %s\n", node.Path)

	return true
}

func (b *AppBuild) _RunBuild() error {
	// find target
	// 1. from commandline
	// 2. from default rule
	// 3. root node which don't have out edge
	targets := b.getTargets()
	if len(targets) == 0 {
		fmt.Printf("no such targets: %s\n", b.Option.Targets)
		return nil
	}

	fmt.Printf("Targets: [ ")
	for _, t := range targets {
		fmt.Printf("%s ", t.Path)
	}
	fmt.Printf("]\n")

	for _, t := range targets {
		var stack []*Node
		b.CollectDitryNodes(t, stack)
		if t.InEdge != nil || !t.InEdge.OutPutReady {
			err := b.Runner.AddTarget(t, nil, 0)
			if err != nil {
				fmt.Printf("AddTarget failed: %v\n", err)
				return err
			}
		}
	}

	b.Runner.Start()

	return nil
}
