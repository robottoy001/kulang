package main

type Node struct {
	Path          string
	InEdge        *Edge
	OutEdges      []*Edge
	ValidOutEdges []*Edge
}

type Edge struct {
	Ins         []VarString
	Outs        []VarString
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
