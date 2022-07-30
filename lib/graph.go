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
package lib

import (
	"fmt"
	"strings"
	"time"

	"gitee.com/kulang/utils"
)

const (
	VisitedNone uint8 = iota
	VisitedInStack
	VisitedDone
)

type NodeStatus struct {
	Dirty bool
	Exist utils.ExistenceStatus
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
	Id          int
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

	StartTime int64

	depsLoaded   bool
	command      string
	hasEvalInOut bool
}

func NewNode(path string) *Node {
	return &Node{
		Path:          path,
		InEdge:        nil,
		OutEdges:      []*Edge{},
		ValidOutEdges: []*Edge{},
		Status: NodeStatus{
			Dirty: false,
			Exist: utils.ExistenceStatusUnknown,
			MTime: time.Time{},
		},
	}
}

func NewEdge(rule *Rule) *Edge {
	return &Edge{
		Id:            utils.GetId(utils.EdgeSlot),
		Rule:          rule,
		Pool:          NewPool("default", 0),
		Scope:         NewScope(nil),
		Outs:          []*Node{},
		Ins:           []*Node{},
		Validations:   []*Node{},
		OutPutReady:   false,
		VisitStatus:   VisitedNone,
		ImplicitOuts:  0,
		ImplicitDeps:  0,
		OrderOnlyDeps: 0,
		StartTime:     0,
		depsLoaded:    false,
		hasEvalInOut:  false,
	}
}

func (e *Edge) String() string {
	s := fmt.Sprintf("\x1B[31mBUILD\x1B[0m %s: %s", e.Outs[0].Path, e.Rule.Name)
	return s
}

func (e *Edge) IsImplicit(index int) bool {
	return index >= (len(e.Ins)-e.ImplicitDeps-e.OrderOnlyDeps) && !e.IsOrderOnly(index)
}

func (e *Edge) IsOrderOnly(index int) bool {
	return index >= len(e.Ins)-e.OrderOnlyDeps
}

func (e *Edge) AllInputReady() bool {
	for _, in := range e.Ins {
		if in.InEdge != nil && !in.InEdge.OutPutReady {
			return false
		}
	}
	return true
}

func (e *Edge) IsPhony() bool {
	return e.Rule.Type == PhonyRule
}

func (e *Edge) QueryVar(varname string) string {

	// in & out
	if !e.hasEvalInOut {
		e.EvalInOut()
		e.hasEvalInOut = true
	}

	varValue := e.Rule.QueryVar(varname)
	if varValue != nil {
		return varValue.Eval(e.Scope)
	}

	varValue = e.Scope.QueryVar(varname)
	if varValue == nil {
		return ""
	}

	return varValue.Eval(e.Scope)
}

func (e *Edge) EvalInOut() {
	buffer := new(strings.Builder)

	explicit_in_deps := len(e.Ins) - e.ImplicitDeps - e.OrderOnlyDeps
	for i := 0; i < explicit_in_deps; i += 1 {
		buffer.WriteString(e.Ins[i].Path)
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

	e.Scope.AppendVar("in", inStr)
	buffer.Reset()

	explicit_out_deps := len(e.Outs) - e.ImplicitOuts
	for i := 0; i < explicit_out_deps; i += 1 {
		buffer.WriteString(e.Outs[i].Path)
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

	e.Scope.AppendVar("out", outStr)

	//fmt.Printf("out: %s\n", buffer.String())
}

func (e *Edge) EvalCommand() string {
	if e.command == "" {
		command := e.QueryVar("command")
		return command
	}
	return e.command
}

func (e *Node) Stat(fs utils.FileSystem) bool {
	finfo, err := fs.Stat(e.Path)
	if err != nil {
		e.Status.Exist = finfo.Exist
		return false
	}

	e.Status.MTime = finfo.MTime
	e.Status.Exist = finfo.Exist
	return true
}

func (e *Node) StatusKnow() bool {
	return e.Status.Exist != utils.ExistenceStatusUnknown
}

func (e *Node) Exist() bool {
	return e.Status.Exist == utils.ExistenceStatusExist
}

func (e *Node) SetDirty(dirty bool) {
	e.Status.Dirty = dirty
}

func (e *Node) SetPhonyMtime(t time.Time) {
	e.Status.MTime = t
}
