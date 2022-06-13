package main

import (
	"fmt"
	"log"
)

type NinjaParser struct {
	*Parser
}

func (self *NinjaParser) Parse() error {
	fmt.Println("parsing")
	self.Scanner.NextToken()
	tok := self.Scanner.GetToken()

	switch {
	case tok.Type == TOKEN_DEFUALT:
		self.parseDefault()
		break
	case tok.Type == TOKEN_BUILD:
		self.parseBuild()
		break
	case tok.Type == TOKEN_RULE:
		self.parseRule()
		break
	case tok.Type == TOKEN_IDENT:
		varName, varValue := self.parseVariable()
		self.Scope.AppendVar(varName, varValue.Eval(self.Scope))
		break
	}

	if tok.Literal != "" {
		fmt.Println(tok.Literal)
	}
	fmt.Println("parser, finished")
	return nil
}

func (self *NinjaParser) parseDefault() {
	// output list
	for {
		out, err := self.Scanner.ScanVarValue(true)
		if len(out.Str) == 0 || err != nil {
			break
		}
		path := out.Eval(self.Scope)
		self.App.AddDefaults(path)
	}
}

func (self *NinjaParser) parseBuild() {
	// out

	// :
	// rule name
	// input list
	// |
	// implict dependence
	// ||
	// order dependence
	// |@
	// validation
}

func (self *NinjaParser) parseRule() {
	// rule name
	// expected TOKEN_IDENT
	self.Scanner.NextToken()
	tok := self.Scanner.GetToken()
	ruleName := tok.Literal

	// expected newline
	self.Scanner.NextToken()
	tok = self.Scanner.GetToken()
	if tok.Type != TOKEN_NEWLINE {
		log.Fatal("expected NEWLINE, got "+tok.Literal, ", type: ", tok.Type)
		return
	}

	rule := NewRule(ruleName)

	// add variable in Rule
	for {
		self.Scanner.NextToken()
		tok := self.Scanner.GetToken()
		if tok.Type != TOKEN_IDENT {
			break
		}
		varName, varValue := self.parseVariable()
		rule.AppendVar(varName, varValue)
	}

	self.Scope.AppendRule(ruleName, rule)
}

func (self *NinjaParser) parseVariable() (string, VarString) {
	tok := self.Scanner.GetToken()
	varName := tok.Literal

	// expectd '='
	self.Scanner.NextToken()
	tok = self.Scanner.GetToken()
	if tok.Type != TOKEN_EQUALS {
		log.Fatal("Expected =, Got ", tok.Literal)
		return "", VarString{}
	}

	varValue, err := self.parseVarValue()
	if err != nil {
		log.Fatal("Parse Value fail")
		return varName, VarString{}
	}

	return varName, varValue
}

func (self *NinjaParser) parseVarValue() (VarString, error) {
	valString, err := self.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}
