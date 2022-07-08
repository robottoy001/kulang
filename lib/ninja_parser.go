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
package lib

import (
	"fmt"
	"log"
	"path"
	"strconv"
)

type NinjaParser struct {
	*Parser
}

func (p *NinjaParser) Parse() error {
	//start := time.Now()
Loop:
	for {
		tok := p.Scanner.NextToken()

		switch {
		case tok.Type == TokenSubninja:
			p.parseInclude(true)
			break
		case tok.Type == TokenInclude:
			p.parseInclude(false)
		case tok.Type == TokenDefault:
			p.parseDefault()
			break
		case tok.Type == TokenBuild:
			p.parseBuild()
			break
		case tok.Type == TokenRule:
			p.parseRule()
			break
		case tok.Type == TokenPool:
			p.parsePool()
			break
		case tok.Type == TokenIdent:
			p.Scanner.BackwardToken()
			//fmt.Printf("INDENT, start\n")
			varName, varValue := p.parseVariable()
			p.Scope.AppendVar(varName, varValue)
			//fmt.Printf("INDENT END %s = %s\n", varName, varValue.Eval(p.Scope))
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

func (p *NinjaParser) parseInclude(new_scope bool) {
	varString, err := p.parseVarValue()
	if err != nil {
		log.Fatal("parseNinja fail", err)
		return
	}
	relative_path := varString.Eval(p.Scope)

	// new parser with new scope
	scope := p.Scope
	if new_scope {
		scope = NewScope(p.Scope)
	}
	subParser := NewParser(p.App, scope)

	// load & parse
	real_path := path.Join(p.App.Option.BuildDir, relative_path)
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

func (p *NinjaParser) parseDefault() {
	// output list
	for {
		out, err := p.Scanner.ScanVarValue(true)
		if len(out.Str) == 0 || err != nil {
			break
		}
		target := out.Eval(p.Scope)
		p.App.AddDefaults(target)
	}
}

func (p *NinjaParser) parseBuild() {
	//fmt.Printf("ParseBuild-----------begin--\n")
	// out
	var outs []*VarString
	for {
		out, _ := p.parsePath()
		if out.Empty() {
			break
		}
		outs = append(outs, out)
	}
	//fmt.Println(outs)

	// |
	// implict out
	var implicit_outs int = 0
	if p.Scanner.PeekToken(TokenPipe) {
		for {
			out, _ := p.parsePath()
			if out.Empty() {
				break
			}
			outs = append(outs, out)
			implicit_outs += 1
		}
	}
	//fmt.Println("implict Out: ", imOuts)

	// :
	if !p.Scanner.PeekToken(TokenColon) {
		log.Fatal("Expected (:), Got Token Type: ", p.Scanner.GetToken().Type)
		return
	}

	// rule name
	if !p.Scanner.PeekToken(TokenIdent) {
		log.Fatal("Expected (TokenIdent), Got Token Type", p.Scanner.GetToken().Type)
		return
	}
	ruleName := p.Scanner.GetToken().Literal
	//fmt.Printf("edge - ruleName %s\n", ruleName)

	// input list
	var ins []*VarString
	for {
		in, _ := p.parsePath()
		if in.Empty() {
			break
		}
		ins = append(ins, in)
	}
	//fmt.Println("ins: ", imOuts)
	// |
	// implicit dependence
	var implicit_deps int = 0
	if p.Scanner.PeekToken(TokenPipe) {
		for {
			in, _ := p.parsePath()
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
	if p.Scanner.PeekToken(TokenPipe2) {
		for {
			in, _ := p.parsePath()
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
	if p.Scanner.PeekToken(TokenPipeAt) {
		for {
			in, _ := p.parsePath()
			if in.Empty() {
				break
			}
			valids = append(valids, in)
		}
	}
	//fmt.Println("valids ins: ", valids)

	// new line
	// expected newline
	if !p.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", p.Scanner.GetToken().Literal,
			",type: ", p.Scanner.GetToken().Type,
			",LineNo: ", p.Scanner.GetToken().Loc.LineNo,
			",Col: ", p.Scanner.GetToken().Loc.Start-p.Scanner.GetToken().Loc.LineStart)
		return
	}

	// new Edge
	//fmt.Println(p.Scope)
	var rule *Rule
	if ruleName == PhonyRule.Name {
		rule = PhonyRule
	} else {
		rule = p.Scope.QueryRule(ruleName)
		if rule == nil {
			log.Fatal("Rule: ", ruleName, " doesn't exist")
			return
		}
	}

	edge := NewEdge(rule)

	edge.ImplicitOuts = implicit_outs
	edge.ImplicitDeps = implicit_deps
	edge.OrderOnlyDeps = order_only_deps

	p.App.AddBuild(edge)
	// variable
	// add variable to edge
	scope := NewScope(p.Scope)

	for p.Scanner.PeekToken(TokenIndent) {
		varName, varValue := p.parseVariable()
		scope.AppendVar(varName, varValue)
	}
	edge.Scope = scope

	// check pool if exist
	poolName := edge.Scope.QueryVar("pool")
	if poolName != nil {
		pool := p.App.FindPool(poolName.Eval(scope))
		edge.Pool = pool
	}

	for _, o := range outs {
		p.App.AddOut(edge, o.Eval(scope))
	}

	for _, i := range ins {
		p.App.AddIn(edge, i.Eval(scope))
	}

	//edge.EvalInOut()

}

func (p *NinjaParser) parseRule() {
	// rule name
	// expected TokenIdent
	tok := p.Scanner.ScanIdent()
	//fmt.Printf("ruleName: %s\n", ruleName)
	//log.Println("Rule, got ", p.Scanner.GetToken().Literal,
	//	",type: ", p.Scanner.GetToken().Type,
	//	",LineNo: ", p.Scanner.GetToken().Loc.LineNo,
	//	",Col: ", p.Scanner.GetToken().Loc.Start-p.Scanner.GetToken().Loc.LineStart)

	// expected newline
	if !p.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", p.Scanner.GetToken().Literal,
			",type: ", p.Scanner.GetToken().Type,
			",LineNo: ", p.Scanner.GetToken().Loc.LineNo,
			",Col: ", p.Scanner.GetToken().Loc.Start-p.Scanner.GetToken().Loc.LineStart)
		return
	}

	rule := NewRule(tok.Literal)

	// add variable in Rule
	for p.Scanner.PeekToken(TokenIndent) {
		varName, varValue := p.parseVariable()
		rule.AppendVar(varName, varValue)
	}

	p.Scope.AppendRule(tok.Literal, rule)
	//fmt.Printf("%v\n", p.Scope.Rules)
}

func (p *NinjaParser) parseVariable() (string, *VarString) {
	//fmt.Printf("--------Begin parse Variable--------------\n")
	tok := p.Scanner.ScanIdent()
	varName := tok.Literal

	// expectd '='
	if !p.Scanner.PeekToken(TokenEquals) {
		log.Fatal("Expected =, Got", tok.Literal)
		return "", &VarString{}
	}

	varValue, err := p.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal("Parse Value fail")
		return varName, &VarString{}
	}
	//fmt.Printf("--------End parse Variable--------------\n")

	return varName, varValue
}

func (p *NinjaParser) parsePath() (*VarString, error) {
	valString, err := p.Scanner.ScanVarValue(true)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}

func (p *NinjaParser) parseVarValue() (*VarString, error) {
	valString, err := p.Scanner.ScanVarValue(false)
	if err != nil {
		log.Fatal(err)
	}

	return valString, err
}

func (p *NinjaParser) parsePool() {
	tok := p.Scanner.NextToken()
	poolName := tok.Literal

	// expected newline
	if !p.Scanner.PeekToken(TokenNewLine) {
		log.Panicln("expected NEWLINE, got ", p.Scanner.GetToken().Literal,
			",type: ", p.Scanner.GetToken().Type,
			",LineNo: ", p.Scanner.GetToken().Loc.LineNo,
			",Col: ", p.Scanner.GetToken().Loc.Start-p.Scanner.GetToken().Loc.LineStart)
		return
	}

	// add variable in Rule
	for p.Scanner.PeekToken(TokenIndent) {
		varName, varValue := p.parseVariable()
		if varName == "depth" {
			//depth, _ := strconv.ParseUint(varValue.Eval(p.Scope), 10, 32)
			depth, _ := strconv.Atoi(varValue.Eval(p.Scope))
			//fmt.Printf("Got POOL depth = %d\n", depth)
			p.App.AddPool(poolName, depth)
		}
	}

}
