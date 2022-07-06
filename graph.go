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

func (self *Edge) QueryVar(varname string) string {

	// in & out
	varValue := self.Rule.QueryVar(varname)
	if varValue != nil {
		return varValue.Eval(self.Scope)
	}

	varValue = self.Scope.QueryVar(varname)
	if varValue == nil {
		return ""
	}

	return varValue.Eval(self.Scope)
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
	command := self.QueryVar("command")
	//v := command.Eval(self.Scope)
	return command
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
