package main

import (
	"log"
	"testing"
)

func TestParseRule(t *testing.T) {
	scope := &Scope{
		Rules: make(map[string]*Rule),
		Vars: map[string]string{
			"in":  "bar.cc",
			"out": "bar.o",
		},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app)
	//case1 := "rule FOO\n    command = gcc $in $out\ndescription = build by gcc"

	err := parser.Load("./testdata/rule1.ninja")
	if err != nil {
		t.FailNow()
	}

	parser.Parse()

	if _, ok := scope.Rules["FOO"]; !ok {
		t.FailNow()
	}
}

func TestParseDefaults(t *testing.T) {
	scope := &Scope{
		Rules: make(map[string]*Rule),
		Vars: map[string]string{
			"in":  "bar.cc",
			"out": "bar.o",
		},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app)
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
		Vars: map[string]string{
			"in":  "bar.cc",
			"out": "bar.o",
		},
		Parent: nil,
	}

	option := &BuildOption{}
	app := NewAppBuild(option)
	app.Scope = scope

	parser := NewParser(app)
	//case1 := "rule FOO\n    command = gcc $in $out\ndescription = build by gcc"

	err := parser.Load("./testdata/build.ninja")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}

	parser.Parse()

}
