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

type Scope struct {
	Rules  map[string]*Rule
	Vars   map[string]*VarString
	Parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Rules:  make(map[string]*Rule),
		Vars:   make(map[string]*VarString),
		Parent: parent,
	}
}

func (self *Scope) QueryVar(k string) *VarString {
	if v, ok := self.Vars[k]; ok {
		return v
	}

	if self.Parent != nil {
		return self.Parent.QueryVar(k)
	}
	return &VarString{Str: []_VarString{}}
}

func (self *Scope) AppendVar(k string, v *VarString) {
	self.Vars[k] = v
}

func (self *Scope) AppendRule(ruleName string, rule *Rule) {
	self.Rules[ruleName] = rule
}

func (self *Scope) QueryRule(ruleName string) *Rule {
	if rule, ok := self.Rules[ruleName]; ok {
		return rule
	}

	if self.Parent != nil {
		return self.Parent.QueryRule(ruleName)
	}
	return nil
}
