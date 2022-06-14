package main

type Pool struct {
	Name  string
	Depth uint32
}

func NewPool(name string, depth uint32) *Pool {
	return &Pool{
		Name:  name,
		Depth: depth,
	}
}
