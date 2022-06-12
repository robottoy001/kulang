package main

import "gitee.com/kulang/build"

type Scope struct {
	Rule   map[string]*build.Rule
	Parent *Scope
	Vars   map[string]string
}

func (self *Scope) QueryVar(k string) string {
	if v, ok := self.Vars[k]; ok {
		return v
	}
	return ""
}
