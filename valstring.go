package main

import "fmt"

type StrType uint8

const (
	ORGINAL StrType = iota
	VARIABLE
)

type _VarString struct {
	Str  string
	Type StrType
}

type VarString struct {
	Str []_VarString
}

func (self *VarString) Append(s string, st StrType) {
	self.Str = append(self.Str, _VarString{Str: s, Type: st})
}

func (self *VarString) Eval(scope *Scope) string {
	var Value string

	for _, v := range self.Str {
		if v.Type == ORGINAL {
			Value += v.Str
		} else if v.Type == VARIABLE {
			Value += scope.QueryVar(v.Str)
		}
	}
	return Value
}

func (self *VarString) String() string {
	var Value string
	for _, v := range self.Str {
		if v.Type == ORGINAL {
			Value += v.Str
		} else if v.Type == VARIABLE {
			Value += fmt.Sprintf("${%s}", v.Str)
			fmt.Println(v.Str)
		}
	}
	return Value
}
