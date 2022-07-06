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
	"log"
	"testing"
)

func TestParseRule(t *testing.T) {
	scope := &Scope{
		Rules: make(map[string]*Rule),
		Vars: map[string]*VarString{
			"in": &VarString{
				Str: []_VarString{
					{
						Str:  "bar.cc",
						Type: ORGINAL,
					},
				}},
			"out": &VarString{
				Str: []_VarString{
					{
						Str:  "bar,o",
						Type: ORGINAL},
				},
			}},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app, scope)
	//case1 := "rule FOO\n    command = gcc $in $out\ndescription = build by gcc"

	err := parser.Load("./testdata/rule1.ninja")
	if err != nil {
		t.FailNow()
	}

	parser.Parse()

	fmt.Println(scope.Rules)

	if _, ok := scope.Rules["FOO"]; !ok {
		fmt.Println(ok)
		t.FailNow()
	}
}

func TestParseDefaults(t *testing.T) {
	scope := &Scope{
		Rules: make(map[string]*Rule),
		Vars: map[string]*VarString{
			"in": &VarString{
				Str: []_VarString{
					{
						Str:  "bar.cc",
						Type: ORGINAL,
					},
				}},
			"out": &VarString{
				Str: []_VarString{
					{
						Str:  "bar,o",
						Type: ORGINAL},
				},
			}},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app, scope)
	//case1 := "rule FOO\n    command = gcc $in $out\ndescription = build by gcc"

	err := parser.Load("./testdata/default.ninja")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	parser.Parse()

	if len(app.Defaults) != 4 {
		log.Printf("Excepted %d nodes, got %d", 4, len(app.Defaults))
		t.FailNow()
	}

}

func TestParseBuild(t *testing.T) {
	scope := &Scope{
		Rules: map[string]*Rule{
			"cxx": &Rule{
				Name: "cxx",
			},
		},
		Vars: map[string]*VarString{
			"in": &VarString{
				Str: []_VarString{
					{
						Str:  "bar.cc",
						Type: ORGINAL,
					},
				}},
			"out": &VarString{
				Str: []_VarString{
					{
						Str:  "bar,o",
						Type: ORGINAL},
				},
			}},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app, scope)
	//case1 := "rule FOO\n    command = gcc $in $out\ndescription = build by gcc"

	err := parser.Load("./testdata/build.ninja")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	parser.Parse()

}
