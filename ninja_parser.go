package main

import (
	"fmt"
	"log"
	"path"
	"strconv"
)

type NinjaParser struct {
	*Parser
}

func (self *NinjaParser) Parse() error {
	//start := time.Now()
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
		case tok.Type == TOKEN_POOL:
			self.parsePool()
			break
		case tok.Type == TOKEN_IDENT:
			self.Scanner.BackwardToken()
			//fmt.Printf("INDENT, start\n")
			varName, varValue := self.parseVariable()
			self.Scope.AppendVar(varName, varValue)
			//fmt.Printf("INDENT END %s = %s\n", varName, varValue.Eval(self.Scope))
			break
		case tok.Type == TOKEN_NEWLINE:
			//	fmt.Printf("NewLINE in Parse\n")
			break
		case tok.Type == TOKEN_EOF:
			//fmt.Printf("Got EOF??\n")
			break Loop
		case tok.Type == TOKEN_INDENT:
			break
		default:
			err := fmt.Errorf("Parse: Unexpected token, type: %d, LineNo: %d, Col: %d\n", tok.Type, tok.Loc.LineNo, tok.Loc.Start-tok.Loc.LineStart)
			return err
		}
	}

	//fmt.Printf(": time elapse %d ms\n", time.Since(start).Milliseconds())
	return nil
}

func (self *NinjaParser) parseInclude(new_scope bool) {
	varString, err := self.parseVarValue()
	if err != nil {
		log.Fatal("parseNinja fail", err)
		return
	}
	relative_path := varString.Eval(self.Scope)

	// new parser with new scope
	scope := self.Scope
	if new_scope {
		scope = NewScope(self.Scope)
	}
	subParser := NewParser(self.App, scope)

	// load & parse
	real_path := path.Join(self.App.Option.BuildDir, relative_path)
	err = subParser.Load(real_path)
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
		target := out.Eval(self.Scope)
		self.App.AddDefaults(target)
	}
}

func (self *NinjaParser) parseBuild() {
	//fmt.Printf("ParseBuild-----------begin--\n")
	// out
	var outs []*VarString
	for {
		out, _ := self.parsePath()
		if out.Empty() {
			break
		}
		outs = append(outs, out)
	}
	//fmt.Println(outs)

	// |
	// implict out
	var implicit_outs int = 0
	if self.Scanner.PeekToken(TOKEN_PIPE) {
		for {
			out, _ := self.parsePath()
			if out.Empty() {
				break
			}
			outs = append(outs, out)
			implicit_outs += 1
		}
	}
	//fmt.Println("implict Out: ", imOuts)

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
	//fmt.Printf("edge - ruleName %s\n", ruleName)

	// input list
	var ins []*VarString
	for {
		in, _ := self.parsePath()
		if in.Empty() {
			break
		}
		ins = append(ins, in)
	}
	//fmt.Println("ins: ", imOuts)
	// |
	// implicit dependence
	var implicit_deps int = 0
	if self.Scanner.PeekToken(TOKEN_PIPE) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			ins = append(ins, in)
			implicit_deps += 1
		}
	}
	//fmt.Printf("implicit ins: %v\n", imIns)

	// ||
	// order dependence
	var order_only_deps int = 0
	if self.Scanner.PeekToken(TOKEN_PIPE2) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			ins = append(ins, in)
			order_only_deps += 1
		}
	}
	//fmt.Printf("implicit ins: %v\n", orIns)

	// |@
	// validation
	var valids []*VarString
	if self.Scanner.PeekToken(TOKEN_PIPEAT) {
		for {
			in, _ := self.parsePath()
			if in.Empty() {
				break
			}
			valids = append(valids, in)
		}
	}
	//fmt.Println("valids ins: ", valids)

	// new line
	// expected newline
	if !self.Scanner.PeekToken(TOKEN_NEWLINE) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	// new Edge
	//fmt.Println(self.Scope)
	rule := self.Scope.QueryRule(ruleName)
	if rule == nil {
		log.Fatal("Rule: ", ruleName, " doesn't exist")
		return
	}

	edge := NewEdge(rule)

	edge.ImplicitOuts = implicit_outs
	edge.ImplicitDeps = implicit_deps
	edge.OrderOnlyDeps = order_only_deps

	self.App.AddBuild(edge)
	// variable
	// add variable to edge
	scope := NewScope(self.Scope)

	for self.Scanner.PeekToken(TOKEN_INDENT) {
		varName, varValue := self.parseVariable()
		scope.AppendVar(varName, varValue)
	}
	edge.Scope = scope

	// check pool if exist
	poolName := edge.Scope.QueryVar("pool")
	if poolName != nil {
		pool := self.App.FindPool(poolName.Eval(scope))
		edge.Pool = pool
	}

	for _, o := range outs {
		self.App.AddOut(edge, o.Eval(scope))
	}

	for _, i := range ins {
		self.App.AddIn(edge, i.Eval(scope))
	}

	//if !edge.IsPhony() {
	//	edge.EvalCommand()
	//}
}

func (self *NinjaParser) parseRule() {
	// rule name
	// expected TOKEN_IDENT
	tok := self.Scanner.ScanIdent()
	//fmt.Printf("ruleName: %s\n", ruleName)
	//log.Println("Rule, got ", self.Scanner.GetToken().Literal,
	//	",type: ", self.Scanner.GetToken().Type,
	//	",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
	//	",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)

	// expected newline
	if !self.Scanner.PeekToken(TOKEN_NEWLINE) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	rule := NewRule(tok.Literal)

	// add variable in Rule
	for self.Scanner.PeekToken(TOKEN_INDENT) {
		varName, varValue := self.parseVariable()
		rule.AppendVar(varName, varValue)
	}

	self.Scope.AppendRule(tok.Literal, rule)
	//fmt.Printf("%v\n", self.Scope.Rules)
}

func (self *NinjaParser) parseVariable() (string, *VarString) {
	//fmt.Printf("--------Begin parse Variable--------------\n")
	tok := self.Scanner.ScanIdent()
	varName := tok.Literal

	// expectd '='
	if !self.Scanner.PeekToken(TOKEN_EQUALS) {
		log.Fatal("Expected =, Got", tok.Literal)
		return "", &VarString{}
	}

	varValue, err := self.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal("Parse Value fail")
		return varName, &VarString{}
	}
	//fmt.Printf("--------End parse Variable--------------\n")

	return varName, varValue
}

func (self *NinjaParser) parsePath() (*VarString, error) {
	valString, err := self.Scanner.ScanVarValue(true)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}

func (self *NinjaParser) parseVarValue() (*VarString, error) {
	valString, err := self.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}

func (self *NinjaParser) parsePool() {
	tok := self.Scanner.NextToken()
	poolName := tok.Literal

	// expected newline
	if !self.Scanner.PeekToken(TOKEN_NEWLINE) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	// add variable in Rule
	for self.Scanner.PeekToken(TOKEN_INDENT) {
		varName, varValue := self.parseVariable()
		if varName == "depth" {
			//depth, _ := strconv.ParseUint(varValue.Eval(self.Scope), 10, 32)
			depth, _ := strconv.Atoi(varValue.Eval(self.Scope))
			//fmt.Printf("Got POOL depth = %d\n", depth)
			self.App.AddPool(poolName, depth)
		}
	}

}
