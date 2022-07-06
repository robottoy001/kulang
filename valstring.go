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
	"bytes"
	"fmt"
)

type StrType uint8

const (
	ORGINAL StrType = iota
	VARIABLE
)

type _VarString struct {
	Str  string
	Type StrType
}

type VarString struct {
	Str []_VarString
}

func (self *VarString) Append(s string, st StrType) {
	self.Str = append(self.Str, _VarString{Str: s, Type: st})
}

func (self *VarString) Eval(scope *Scope) string {
	var Value bytes.Buffer

	for _, v := range self.Str {
		if v.Type == ORGINAL {
			Value.WriteString(v.Str)
		} else if v.Type == VARIABLE {
			Value.WriteString(scope.QueryVar(v.Str).Eval(scope))
		}
	}
	return Value.String()
}

func (self *VarString) String() string {
	var Value string
	for _, v := range self.Str {
		if v.Type == ORGINAL {
			Value += v.Str
		} else if v.Type == VARIABLE {
			Value += fmt.Sprintf("${%s}", v.Str)
			fmt.Println(v.Str)
		}
	}
	return Value
}

func (self *VarString) Empty() bool {
	return len(self.Str) == 0
}
