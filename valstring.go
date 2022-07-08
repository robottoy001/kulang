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

func (v *VarString) Append(s string, st StrType) {
	v.Str = append(v.Str, _VarString{Str: s, Type: st})
}

func (v *VarString) Eval(scope *Scope) string {
	var Value strings.Builder

	for _, v := range v.Str {
		if v.Type == ORGINAL {
			Value.WriteString(v.Str)
		} else if v.Type == VARIABLE {
			Value.WriteString(scope.QueryVar(v.Str).Eval(scope))
		}
	}
	return Value.String()
}

func (v *VarString) String() string {
	var value strings.Builder
	for _, v := range v.Str {
		if v.Type == ORGINAL {
			value.WriteString(v.Str)
		} else if v.Type == VARIABLE {
			value.WriteString(fmt.Sprintf("${%s}", v.Str))
			fmt.Println(v.Str)
		}
	}
	return value.String()
}

func (v *VarString) Empty() bool {
	return len(v.Str) == 0
}
