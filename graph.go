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
	Outs        []*Node
	Ins         []*Node
	Validations []*Node
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
		Outs:        []*Node{},
		Ins:         []*Node{},
		Validations: []*Node{},
	}
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
