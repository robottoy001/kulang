package main

type Rule struct {
	Name string
	Vars map[string]*VarString
}

var PhonyRule = &Rule{
	Name: "phony",
	Vars: map[string]*VarString{},
}

func NewRule(name string) *Rule {
	return &Rule{
		Name: name,
		Vars: make(map[string]*VarString),
	}
}

func (self *Rule) AppendVar(k string, v *VarString) {
	self.Vars[k] = v
}

func (self *Rule) QueryVar(k string) *VarString {
	if v, ok := self.Vars[k]; ok {
		return v
	}
	return nil
}
