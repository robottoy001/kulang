package main

type Rule struct {
	Name string
	Vars map[string]VarString
}

func NewRule(name string) *Rule {
	return &Rule{
		Name: name,
		Vars: make(map[string]VarString),
	}
}

func (self *Rule) AppendVar(k string, v VarString) {
	self.Vars[k] = v
}
