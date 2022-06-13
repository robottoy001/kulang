package main

type Scope struct {
	Rules  map[string]*Rule
	Vars   map[string]string
	Parent *Scope
}

func (self *Scope) QueryVar(k string) string {
	if v, ok := self.Vars[k]; ok {
		return v
	}
	return ""
}

func (self *Scope) AppendVar(k string, v string) {
	self.Vars[k] = v
}

func (self *Scope) AppendRule(ruleName string, rule *Rule) {
	self.Rules[ruleName] = rule
}
