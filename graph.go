package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type ExistenceStatus uint8

const (
	ExistenceStatusUnknown ExistenceStatus = iota
	ExistenceStatusMissing
	ExistenceStatusExist
)

type NodeStatus struct {
	Dirty bool
	Exist ExistenceStatus
	MTime time.Time
}

type Node struct {
	Path          string
	InEdge        *Edge
	OutEdges      []*Edge
	ValidOutEdges []*Edge
	Status        NodeStatus
}

type Edge struct {
	Rule        *Rule
	Pool        *Pool
	Scope       *Scope
	Outs        []*Node
	Ins         []*Node
	Validations []*Node
	OutPutReady bool
}

func NewNode(path string) *Node {
	return &Node{
		Path:          path,
		InEdge:        nil,
		OutEdges:      []*Edge{},
		ValidOutEdges: []*Edge{},
		Status: NodeStatus{
			Dirty: false,
			Exist: ExistenceStatusUnknown,
			MTime: time.Time{},
		},
	}
}

func NewEdge(rule *Rule) *Edge {
	return &Edge{
		Rule:        rule,
		Pool:        NewPool("default", 0),
		Scope:       NewScope(nil),
		Outs:        []*Node{},
		Ins:         []*Node{},
		Validations: []*Node{},
		OutPutReady: false,
	}
}

func (self *Edge) AllInputReady() bool {
	for _, in := range self.Ins {
		if in.InEdge != nil && !in.InEdge.OutPutReady {
			return false
		}
	}
	return true
}

func (self *Edge) IsPhony() bool {
	return self.Rule.Name == "phony"
}

func (self *Edge) EvalInOut() {
	buffer := new(strings.Builder)

	for _, in := range self.Ins {
		buffer.WriteString(in.Path)
		buffer.WriteString(" ")
	}

	self.Scope.AppendVar("in", strings.TrimRight(buffer.String(), " "))
	buffer.Reset()

	for _, out := range self.Outs {
		buffer.WriteString(out.Path)
		buffer.WriteString(" ")
	}
	self.Scope.AppendVar("out", strings.TrimRight(buffer.String(), " "))

	//fmt.Printf("out: %s\n", buffer.String())
}

func (self *Edge) EvalCommand() string {
	command := self.Rule.QueryVar("command")
	self.EvalInOut()
	v := command.Eval(self.Scope)
	fmt.Printf("%s\n", v)
	return v
}

func (self *Node) Stat() bool {
	finfo, err := os.Stat(self.Path)
	if err != nil {
		if os.IsExist(err) {
			self.Status.Exist = ExistenceStatusExist
		} else if os.IsNotExist(err) {
			self.Status.Exist = ExistenceStatusMissing
		} else {
			fmt.Printf("%s: %v", self.Path, err)
			return false
		}
		// else unknown
		return true
	}

	self.Status.MTime = finfo.ModTime()
	self.Status.Exist = ExistenceStatusExist
	return true
}

func (self *Node) StatusKnow() bool {
	return self.Status.Exist != ExistenceStatusUnknown
}

func (self *Node) Exist() bool {
	return self.Status.Exist == ExistenceStatusExist
}

func (self *Node) SetDirty(dirty bool) {
	self.Status.Dirty = dirty
}
