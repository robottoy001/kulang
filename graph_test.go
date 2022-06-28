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
