package main

type Rule struct {
	Name string
	Vars map[string]string
}

func NewRule(name string) *Rule {
	return &Rule{
		Name: name,
		Vars: make(map[string]string),
	}
}

func (self *Rule) AppendVar(k string, v string) {
	self.Vars[k] = v
}
