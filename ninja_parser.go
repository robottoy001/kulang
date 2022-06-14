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

Loop:
	for {
		tok := self.Scanner.NextToken()

		switch {
		case tok.Type == TOKEN_SUBNINJA:
			self.parseInclude(true)
			break
		case tok.Type == TOKEN_INCLUDE:
			self.parseInclude(false)
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
		case tok.Type == TOKEN_NEWLINE:
			break
		case tok.Type == TOKEN_EOF:
			break Loop
		default:
			err := fmt.Errorf("Unexpected token, type: %d, LineNo: %d, Col: %d\n", tok.Type, tok.Loc.LineNo, tok.Loc.Start-tok.Loc.LineStart)
			return err
		}
	}

	fmt.Println("parser, finished")
	return nil
}

func (self *NinjaParser) parseInclude(new_scope bool) {
	varString, err := self.parseVarValue()
	if err != nil {
		log.Fatal("parseNinja fail", err)
		return
	}
	path := varString.Eval(self.Scope)

	// new parser with new scope
	app := self.App
	if new_scope {
		app.Scope = NewScope(self.App.Scope)
	}
	subParser := NewParser(app)

	// load & parse
	err = subParser.Load(path)
	if err != nil {
		log.Fatal("parseSubNinja fail", err)
		return
	}

	err = subParser.Parse()
	if err != nil {
		log.Fatal("parseSubNinja fail", err)
		return
	}
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
	var outs []VarString
	for {
		out, _ := self.parsePath()
		if out.Empty() {
			break
		}
		outs = append(outs, out)
	}
	fmt.Println(outs)
	// |
	// implict out
	var imOuts []VarString
	if self.Scanner.PeekToken(TOKEN_PIPE) {
		for {
			out, _ := self.parsePath()
			if out.Empty() {
				break
			}
			imOuts = append(imOuts, out)
		}
	}
	fmt.Println("implict Out: ", imOuts)

	// :
	if !self.Scanner.PeekToken(TOKEN_COLON) {
		log.Fatal("Expected (:), Got Token Type: ", self.Scanner.GetToken().Type)
		return
	}

	// rule name
	if !self.Scanner.PeekToken(TOKEN_IDENT) {
		log.Fatal("Expected (TOKEN_IDENT), Got Token Type", self.Scanner.GetToken().Type)
		return
	}
	ruleName := self.Scanner.GetToken().Literal
	fmt.Printf("ruleName %s\n", ruleName)

	// input list
	var ins []VarString
	for {
		in, _ := self.parsePath()
		if in.Empty() {
			break
		}
		ins = append(ins, in)
	}
	fmt.Println("ins: ", imOuts)
	// |
	// implicit dependence
	var imIns []VarString
	if self.Scanner.PeekToken(TOKEN_PIPE) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			imIns = append(imIns, in)
		}
	}
	fmt.Println("implicit ins: ", imIns)

	// ||
	// order dependence
	var orIns []VarString
	if self.Scanner.PeekToken(TOKEN_PIPE2) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			orIns = append(orIns, in)
		}
	}
	fmt.Println("implicit ins: ", orIns)

	// |@
	// validation
	var valids []VarString
	if self.Scanner.PeekToken(TOKEN_PIPEAT) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			valids = append(valids, in)
		}
	}
	fmt.Println("valids ins: ", valids)

	// new line
	// expected newline
	if !self.Scanner.PeekToken(TOKEN_NEWLINE) {
		log.Fatal("expected NEWLINE, got "+self.Scanner.GetToken().Literal, ", type: ", self.Scanner.GetToken().Type)
		return
	}

	// new Edge
	fmt.Println(self.Scope)
	rule := self.Scope.QueryRule(ruleName)
	if rule == nil {
		log.Fatal("Rule: ", ruleName, " doesn't exist")
		return
	}

	edge := NewEdge(rule)
	// variable
	// add variable to edge
	scope := NewScope(self.Scope)
	for {
		tok := self.Scanner.NextToken()
		if tok.Type != TOKEN_IDENT {
			break
		}
		varName, varValue := self.parseVariable()
		scope.AppendVar(varName, varValue.Eval(scope))
	}
	edge.Scope = scope
}

func (self *NinjaParser) parseRule() {
	// rule name
	// expected TOKEN_IDENT
	tok := self.Scanner.NextToken()
	ruleName := tok.Literal

	// expected newline
	if !self.Scanner.PeekToken(TOKEN_NEWLINE) {
		log.Fatal("expected NEWLINE, got ", self.Scanner.GetToken().Literal, ", type: ", self.Scanner.GetToken().Type)
		return
	}

	rule := NewRule(ruleName)

	// add variable in Rule
	for {
		tok := self.Scanner.NextToken()
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
	tok = self.Scanner.NextToken()
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

func (self *NinjaParser) parsePath() (VarString, error) {
	valString, err := self.Scanner.ScanVarValue(true)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}

func (self *NinjaParser) parseVarValue() (VarString, error) {
	valString, err := self.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}
