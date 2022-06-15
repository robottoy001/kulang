package main

type Pool struct {
	Name  string
	Depth int
}

func NewPool(name string, depth int) *Pool {
	return &Pool{
		Name:  name,
		Depth: depth,
	}
}
