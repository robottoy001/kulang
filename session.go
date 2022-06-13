package main

type Session struct {
	// pool
	Pool *Pool
	// edges
	Edges []*Edge

	// default
	Scope *Scope
}
