package main

import (
	"fmt"
	"strings"
	"time"
)

type ExistenceStatus uint8

const (
	ExistenceStatusUnknown ExistenceStatus = iota
	ExistenceStatusMissing
	ExistenceStatusExist
)

const (
	VISITED_NONE uint8 = iota
	VISITED_IN_STACK
	VISITED_DONE
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
	VisitStatus uint8

	ImplicitOuts  int
	ImplicitDeps  int
	OrderOnlyDeps int
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
		Rule:          rule,
		Pool:          NewPool("default", 0),
		Scope:         NewScope(nil),
		Outs:          []*Node{},
		Ins:           []*Node{},
		Validations:   []*Node{},
		OutPutReady:   false,
		VisitStatus:   VISITED_NONE,
		ImplicitOuts:  0,
		ImplicitDeps:  0,
		OrderOnlyDeps: 0,
	}
}

func (self *Edge) String() string {
	s := fmt.Sprintf("\x1B[31mBUILD\x1B[0m %s: %s ", self.Outs[0].Path, self.Rule.Name)
	var insStr string
	//	for i := 0; i < len(self.Ins)-self.ImplicitDeps-self.OrderOnlyDeps; i += 1 {
	//		insStr += self.Ins[i].Path
	//		insStr += " "
	//	}
	s += insStr
	return s
}

func (self *Edge) IsImplicit(index int) bool {
	return index >= (len(self.Ins)-self.ImplicitDeps-self.OrderOnlyDeps) && !self.IsOrderOnly(index)
}

func (self *Edge) IsOrderOnly(index int) bool {
	return index >= len(self.Ins)-self.OrderOnlyDeps
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

	explicit_in_deps := len(self.Ins) - self.ImplicitDeps - self.OrderOnlyDeps
	for i := 0; i < explicit_in_deps; i += 1 {
		buffer.WriteString(self.Ins[i].Path)
		buffer.WriteString(" ")
	}

	inStr := &VarString{
		Str: []_VarString{
			{
				Str:  strings.TrimRight(buffer.String(), " "),
				Type: ORGINAL,
			},
		},
	}

	self.Scope.AppendVar("in", inStr)
	buffer.Reset()

	explicit_out_deps := len(self.Outs) - self.ImplicitOuts
	for i := 0; i < explicit_out_deps; i += 1 {
		buffer.WriteString(self.Outs[i].Path)
		buffer.WriteString(" ")
	}

	outStr := &VarString{
		Str: []_VarString{
			{
				Str:  strings.TrimRight(buffer.String(), " "),
				Type: ORGINAL,
			},
		},
	}

	self.Scope.AppendVar("out", outStr)

	//fmt.Printf("out: %s\n", buffer.String())
}

func (self *Edge) EvalCommand() string {
	command := self.Rule.QueryVar("command")
	self.EvalInOut()
	v := command.Eval(self.Scope)
	return v
}

func (self *Node) Stat(fs FileSystem) bool {
	finfo, err := fs.Stat(self.Path)
	if err != nil {
		self.Status.Exist = finfo.Exist
		return false
	}

	self.Status.MTime = finfo.MTime
	self.Status.Exist = finfo.Exist
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

func (self *Node) SetPhonyMtime(t time.Time) {
	self.Status.MTime = t
}
