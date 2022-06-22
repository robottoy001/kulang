package main

type Scope struct {
	Rules  map[string]*Rule
	Vars   map[string]*VarString
	Parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Rules:  make(map[string]*Rule),
		Vars:   make(map[string]*VarString),
		Parent: parent,
	}
}

func (self *Scope) QueryVar(k string) *VarString {
	if v, ok := self.Vars[k]; ok {
		return v
	}

	if self.Parent != nil {
		return self.Parent.QueryVar(k)
	}
	return nil
}

func (self *Scope) AppendVar(k string, v *VarString) {
	self.Vars[k] = v
}

func (self *Scope) AppendRule(ruleName string, rule *Rule) {
	self.Rules[ruleName] = rule
}

func (self *Scope) QueryRule(ruleName string) *Rule {
	if rule, ok := self.Rules[ruleName]; ok {
		return rule
	}

	if self.Parent != nil {
		return self.Parent.QueryRule(ruleName)
	}
	return nil
}
