/*
 * Copyright (c) 2022 Huawei Device Co., Ltd.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
		case tok.Type == TokenSubninja:
			self.parseInclude(true)
			break
		case tok.Type == TokenInclude:
			self.parseInclude(false)
		case tok.Type == TokenDefault:
			self.parseDefault()
			break
		case tok.Type == TokenBuild:
			self.parseBuild()
			break
		case tok.Type == TokenRule:
			self.parseRule()
			break
		case tok.Type == TokenPool:
			self.parsePool()
			break
		case tok.Type == TokenIdent:
			self.Scanner.BackwardToken()
			//fmt.Printf("INDENT, start\n")
			varName, varValue := self.parseVariable()
			self.Scope.AppendVar(varName, varValue)
			//fmt.Printf("INDENT END %s = %s\n", varName, varValue.Eval(self.Scope))
			break
		case tok.Type == TokenNewLine:
			//	fmt.Printf("NewLINE in Parse\n")
			break
		case tok.Type == TokenEof:
			//fmt.Printf("Got EOF??\n")
			break Loop
		case tok.Type == TokenIndent:
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
	if self.Scanner.PeekToken(TokenPipe) {
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
	if !self.Scanner.PeekToken(TokenColon) {
		log.Fatal("Expected (:), Got Token Type: ", self.Scanner.GetToken().Type)
		return
	}

	// rule name
	if !self.Scanner.PeekToken(TokenIdent) {
		log.Fatal("Expected (TokenIdent), Got Token Type", self.Scanner.GetToken().Type)
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
	if self.Scanner.PeekToken(TokenPipe) {
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
	if self.Scanner.PeekToken(TokenPipe2) {
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
	if self.Scanner.PeekToken(TokenPipeAt) {
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
	if !self.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	// new Edge
	//fmt.Println(self.Scope)
	var rule *Rule
	if ruleName == PhonyRule.Name {
		rule = PhonyRule
	} else {
		rule = self.Scope.QueryRule(ruleName)
		if rule == nil {
			log.Fatal("Rule: ", ruleName, " doesn't exist")
			return
		}
	}

	edge := NewEdge(rule)

	edge.ImplicitOuts = implicit_outs
	edge.ImplicitDeps = implicit_deps
	edge.OrderOnlyDeps = order_only_deps

	self.App.AddBuild(edge)
	// variable
	// add variable to edge
	scope := NewScope(self.Scope)

	for self.Scanner.PeekToken(TokenIndent) {
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

	edge.EvalInOut()

}

func (self *NinjaParser) parseRule() {
	// rule name
	// expected TokenIdent
	tok := self.Scanner.ScanIdent()
	//fmt.Printf("ruleName: %s\n", ruleName)
	//log.Println("Rule, got ", self.Scanner.GetToken().Literal,
	//	",type: ", self.Scanner.GetToken().Type,
	//	",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
	//	",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)

	// expected newline
	if !self.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	rule := NewRule(tok.Literal)

	// add variable in Rule
	for self.Scanner.PeekToken(TokenIndent) {
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
	if !self.Scanner.PeekToken(TokenEquals) {
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
	if !self.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", self.Scanner.GetToken().Literal,
			",type: ", self.Scanner.GetToken().Type,
			",LineNo: ", self.Scanner.GetToken().Loc.LineNo,
			",Col: ", self.Scanner.GetToken().Loc.Start-self.Scanner.GetToken().Loc.LineStart)
		return
	}

	// add variable in Rule
	for self.Scanner.PeekToken(TokenIndent) {
		varName, varValue := self.parseVariable()
		if varName == "depth" {
			//depth, _ := strconv.ParseUint(varValue.Eval(self.Scope), 10, 32)
			depth, _ := strconv.Atoi(varValue.Eval(self.Scope))
			//fmt.Printf("Got POOL depth = %d\n", depth)
			self.App.AddPool(poolName, depth)
		}
	}

}
