package main

type Session struct {
	// pool
	Pools map[string]*Pool
	// edges
	Edges []*Edge

	// default
	Scope *Scope
}
