package main

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
	Outs        []VarString
	Ins         []VarString
	Validations []VarString
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
		Outs:        []VarString{},
		Ins:         []VarString{},
		Validations: []VarString{},
	}
}
