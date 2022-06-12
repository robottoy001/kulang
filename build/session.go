package build

type Session struct {
	// pool
	Pool *Pool
	// edges
	Edges []*Edge

	// default
	Scope *Scope
}
