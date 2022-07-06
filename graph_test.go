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
	"log"
	"testing"
)

func TestMissingImplict(t *testing.T) {

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Fs = VirtualFileSystem{Files: map[string]*File{}}

	scope := &Scope{
		Rules: map[string]*Rule{
			"cat": {
				Name: "cat",
			},
		},
	}
	parser := NewParser(app, scope)

	err := parser.Load("./testdata/graph_01.missimplict")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	parser.Parse()

	app.Fs.CreateFile("in", "")
	app.Fs.CreateFile("out", "")

	targets := app.getTargets()
	var stack []*Node
	app.CollectDitryNodes(targets[0], stack)

	node := app.FindNode("out")
	if !node.Status.Dirty {
		t.FailNow()
	}
}
