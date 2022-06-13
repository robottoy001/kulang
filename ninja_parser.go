package main

import (
	"fmt"
	"log"
)

type NinjaParser struct {
	*Parser
}

func (self *NinjaParser) Parse() error {
	self.Scanner.NextToken()
	tok := self.Scanner.GetToken()

	switch {
	case tok.Type == TOKEN_RULE:
		self.parseRule()
		break
	case tok.Type == TOKEN_IDENT:
		varName, varValue := self.parseVariable()
		self.Scope.AppendVar(varName, varValue)
		break
	}
	fmt.Printf("token type: %d\n", tok.Type)
	if tok.Literal != "" {
		fmt.Println(tok.Literal)
	}
	fmt.Println("parsing")
	fmt.Println("parser, finished")
	return nil
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
		log.Fatal("expected NEWLINE, got " + tok.Literal)
		return
	}

	rule := NewRule(ruleName)

	// add variable in Rule
	for {
		self.Scanner.NextToken()
		tok := self.Scanner.GetToken()

		varName, varValue := self.parseVariable()
		rule.AppendVar(varName, varValue)
	}

	self.Scope.AppendRule(ruleName, rule)
}

func (self *NinjaParser) parseVariable() (string, string) {
	tok := self.Scanner.GetToken()
	varName := tok.Literal

	// expectd '='
	self.Scanner.NextToken()
	tok = self.Scanner.GetToken()
	if tok.Type != TOKEN_EQUALS {
		log.Fatal("Expected =, Got ", tok.Literal)
		return "", ""
	}

	varValue, err := self.parseVarValue()
	if err != nil {
		log.Fatal("Parse Value fail")
		return varName, ""
	}

	return varName, varValue
}

func (self *NinjaParser) parseVarValue() (string, error) {
	valString, err := self.Scanner.ScanVarValue()
	if err != nil {
		log.Fatal(err)
	}

	return valString.Eval(self.Scope), err
}
