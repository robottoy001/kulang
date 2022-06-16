package main

import (
	"fmt"
	"strings"
)

type Node struct {
	Path          string
	InEdge        *Edge
	OutEdges      []*Edge
	ValidOutEdges []*Edge
}

type Edge struct {
	Rule        *Rule
	Pool        *Pool
	Scope       *Scope
	Outs        []VarString
	Ins         []VarString
	Validations []VarString
}

func NewNode(path string) *Node {
	return &Node{
		Path:          path,
		InEdge:        nil,
		OutEdges:      []*Edge{},
		ValidOutEdges: []*Edge{},
	}
}

func NewEdge(rule *Rule) *Edge {
	return &Edge{
		Rule:        rule,
		Pool:        NewPool("default", 0),
		Scope:       NewScope(nil),
		Outs:        []VarString{},
		Ins:         []VarString{},
		Validations: []VarString{},
	}
}

func (self *Edge) EvalInOut() {
	buffer := new(strings.Builder)

	for _, in := range self.Ins {
		buffer.WriteString(in.Eval(self.Scope))
	}

	self.Scope.AppendVar("in", buffer.String())
	buffer.Reset()

	for _, out := range self.Outs {
		buffer.WriteString(out.Eval(self.Scope))
	}
	self.Scope.AppendVar("out", buffer.String())

	fmt.Printf("out: %s\n", buffer.String())
}

func (self *Edge) EvalCommand() string {
	command := self.Rule.QueryVar("command")
	self.EvalInOut()
	v := command.Eval(self.Scope)
	fmt.Printf("real command : %s\n", v)
	return v
}
