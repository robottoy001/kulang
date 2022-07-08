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
	"log"
)

const (
	TokenBuild uint8 = iota
	TokenColon
	TokenDefault
	TokenEquals
	TokenIdent
	TokenInclude
	TokenSubninja
	TokenIndent
	TokenNewLine
	TokenPipe
	TokenPipe2
	TokenPipeAt
	TokenPool
	TokenRule
	TokenEof
)

var keyWords = map[string]uint8{
	"build":    TokenBuild,
	"rule":     TokenRule,
	"include":  TokenInclude,
	"subninja": TokenSubninja,
	"pool":     TokenPool,
	"default":  TokenDefault,
}

// totally compatiable with Ninja
type NinjaScanner struct {
	*Scanner
}

func (ns *NinjaScanner) Reset(ct []byte) {
	ns.Content = ct

	tok := Token{
		Type:    TokenEof,
		Literal: "",
		Loc: Location{
			LineNo:    1,
			Start:     0,
			End:       0,
			LineStart: 0,
		},
	}
	ns.LastToken = tok
	ns.Token = tok
	ns.Pos.LineNo = 1
	ns.Pos.Offset = 0
	ns.Pos.LineStart = 0
	ns.buffer.Reset()
}

func (ns *NinjaScanner) isIdentifier(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '.' || ch == '-' {
		return true
	}
	return false
}

func (ns *NinjaScanner) isSimpleIdentifier(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
		return true
	}
	return false
}

func (ns *NinjaScanner) skipWhiteSpace() {
	for {
		ch := ns.peek()
		switch ch {
		case ' ', '\t':
			ns.advance()
			continue
		}
		break
	}
}

func (ns *NinjaScanner) newLine() {
	ns.Pos.LineNo += 1
	ns.Pos.LineStart = ns.Pos.Offset
}

func (ns *NinjaScanner) GetToken() Token {
	ns.Token.Literal = string(ns.Content[ns.Token.Loc.Start:ns.Token.Loc.End])
	return ns.Token
}

func (ns *NinjaScanner) BackwardToken() {
	ns.Token = ns.LastToken
	ns.Pos.LineNo = ns.LastToken.Loc.LineNo
	ns.Pos.Offset = ns.LastToken.Loc.Start
	ns.Pos.LineStart = ns.LastToken.Loc.LineStart
}

func (ns *NinjaScanner) PeekToken(expected uint8) bool {
	nextToken := ns.NextToken()
	if nextToken.Type == expected {
		return true
	}
	ns.BackwardToken()
	return false
}

func (ns *NinjaScanner) tokenStart() {
	ns.Token.Loc.LineNo = ns.Pos.LineNo
	ns.Token.Loc.Start = ns.Pos.Offset
	ns.Token.Loc.LineStart = ns.Pos.LineStart
}

func (ns *NinjaScanner) tokenEnd() {
	ns.Token.Loc.End = ns.Pos.Offset
	ns.LastToken = ns.Token
}

func (ns *NinjaScanner) NextToken() Token {
	for {
		ns.tokenStart()
		ch := ns.peek()
		switch {
		case ch == byte(' ') || ch == byte('\t'):
			// from new line or start of file
			if ns.LastToken.Type == TokenNewLine {
				ns.skipWhiteSpace()
				ns.Token.Type = TokenIndent
				ns.tokenEnd()
				return ns.Token
			}
			ns.skipWhiteSpace()
			continue
		case ch == byte('\r'):
			ns.advance()
			if ns.peek() == byte('\n') {
				ns.advance()
			}
			ns.newLine()
			ns.Token.Type = TokenNewLine
			ns.tokenEnd()
			return ns.Token
		case ch == byte('\n'):
			ns.newLine()
			ns.advance()
			ns.Token.Type = TokenNewLine
			ns.tokenEnd()
			return ns.Token
		case ch == byte('#'):
			ns.advance()
			ns.skipComment()
			continue
		case ch == byte('|'):
			err := ns.scanPipe()
			if err != nil {
				log.Fatal(err)
			}
			ns.tokenEnd()
			return ns.Token
		case ch == ':':
			ns.Token.Type = TokenColon
			ns.advance()
			ns.tokenEnd()
			return ns.Token
		case ch == '=':
			ns.Token.Type = TokenEquals
			ns.advance()
			ns.tokenEnd()
			return ns.Token
			// keywords
			// build, rule,pool, default
			// include, subninja
		case ns.isIdentifier(ch) == true:
			err := ns.scanIdentifier()
			if err != nil {
				log.Fatal(err)
			}
			ns.tokenEnd()
			return ns.Token
		// variable
		case ch == INVALID_CHAR:
			ns.Token.Type = TokenEof
			ns.tokenEnd()
			return ns.Token
		default:
			// error
			log.Panicf("unexpected token (%s) in Line: %d, Col: %d",
				string(ch), ns.Pos.LineNo, ns.Pos.Offset-ns.Pos.LineStart)
			return ns.Token
		}
	}
}

// Iterator
func (ns *NinjaScanner) advance() {
	ns.Pos.Offset += 1
}
func (ns *NinjaScanner) backward() {
	ns.Pos.Offset -= 1
}

func (ns *NinjaScanner) hasNext() bool {
	return ns.Pos.Offset <= uint32((len(ns.Content) - 1))
}

func (ns *NinjaScanner) peek() uint8 {
	if ns.hasNext() {
		return ns.Content[ns.Pos.Offset]
	}
	return INVALID_CHAR
}

func (ns *NinjaScanner) next() uint8 {
	n := ns.peek()
	ns.advance()
	return n
}

func (ns *NinjaScanner) skipComment() {
	for {
		ch := ns.next()
		switch ch {
		case byte('\r'): // \r
			if ns.peek() == byte('\n') {
				ns.advance()
			}
			ns.newLine()
			return
		case '\n':
			ns.newLine()
			return
		}
	} //for
}

func (ns *NinjaScanner) scanPipe() error {
	ns.advance()

	ch := ns.peek()
	switch ch {
	case '@':
		ns.Token.Type = TokenPipeAt
		ns.advance()
		break
	case '|':
		ns.Token.Type = TokenPipe2
		ns.advance()
		break
	default:
		ns.Token.Type = TokenPipe
	}
	return nil
}

func (ns *NinjaScanner) getIdentifierType(identifier string) uint8 {
	if token_type, ok := keyWords[identifier]; ok {
		return token_type
	}

	return TokenIdent
}

func (ns *NinjaScanner) scanIdentifier() error {
	start := ns.Pos.Offset
Loop:
	for {
		ns.advance()
		ch := ns.peek()
		switch {
		case ns.isIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	ns.Token.Type = ns.getIdentifierType(string(ns.Content[start:ns.Pos.Offset]))
	ns.Token.Loc.Start = start
	ns.Token.Loc.End = ns.Pos.Offset
	return nil
}

func (ns *NinjaScanner) scanSimpleIdentifier() error {
	start := ns.Pos.Offset
Loop:
	for {
		ns.advance()
		ch := ns.peek()
		switch {
		case ns.isSimpleIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	ns.Token.Type = ns.getIdentifierType(string(ns.Content[start:ns.Pos.Offset]))
	ns.Token.Loc.Start = start
	ns.Token.Loc.End = ns.Pos.Offset
	return nil
}

func (ns *NinjaScanner) ScanVarValue(path bool) (*VarString, error) {
	ns.skipWhiteSpace()
	var value VarString
	//var tmpStr bytes.Buffer
	var strtype StrType = ORGINAL
	var err error = nil

	ns.tokenStart()
Loop:
	for {
		ch := ns.peek()
		ns.advance()
		switch {
		case ch == '$':
			next := ns.peek()
			// escape
			if next == ':' || next == '$' || next == ' ' {
				//tmpStr.WriteByte(next)
				ns.buffer.WriteByte(next)
				ns.advance()
				break
			}
			// continue
			if next == '\n' {
				ns.advance()
				ns.skipWhiteSpace()
				ns.newLine()
				break
			}
			if next == '\r' {
				ns.advance()
				if ns.peek() == '\n' {
					ns.advance()
				}
				ns.skipWhiteSpace()
				ns.newLine()
				break
			}

			// variable name
			if next == '{' {
				// save normal string if have
				if ns.buffer.Len() != 0 {
					//value.Append(tmpStr.String(), strtype)
					value.Append(ns.buffer.String(), strtype)
				}

				ns.advance()

				ns.buffer.Reset()
				// scan ${varname}
				strtype = VARIABLE
				err := ns.scanIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(string(ns.Content[ns.Token.Loc.Start:ns.Token.Loc.End]), strtype)
			}

			if ns.isSimpleIdentifier(next) {
				if ns.buffer.Len() != 0 {
					//value.Append(tmpStr.String(), strtype)
					value.Append(ns.buffer.String(), strtype)
				}

				ns.buffer.Reset()

				// scan ${varname}
				strtype = VARIABLE
				err := ns.scanSimpleIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(string(ns.Content[ns.Token.Loc.Start:ns.Token.Loc.End]), strtype)
				strtype = ORGINAL
			}
			break
		case ch == '}':
			// reset to original string
			strtype = ORGINAL
			break
		case ch == '\r' || ch == '\n':
			if path {
				ns.backward()
			} else {
				ns.newLine()
			}
			ns.Token.Type = TokenNewLine
			break Loop
		case ch == ' ' || ch == '|' || ch == ':':
			if path {
				ns.backward()
				break Loop
			}
			//tmpStr.WriteByte(ch)
			ns.buffer.WriteByte(ch)
		case ch == INVALID_CHAR:
			break Loop
		default:
			//tmpStr.WriteByte(ch)
			ns.buffer.WriteByte(ch)
		}
	}
	if ns.buffer.Len() != 0 {
		//value.Append(tmpStr.String(), strtype)
		value.Append(ns.buffer.String(), strtype)
	}

	ns.buffer.Reset()

	ns.tokenEnd()

	return &value, err
}

func (ns *NinjaScanner) ScanIdent() Token {
	ns.skipWhiteSpace()
	start := ns.Pos.Offset
	ns.tokenStart()
Loop:
	for {
		ns.advance()
		ch := ns.peek()
		switch {
		case ns.isIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	literal := ns.Content[start:ns.Pos.Offset]
	ns.Token.Type = TokenIdent
	ns.Token.Literal = string(literal)
	ns.tokenEnd()
	//fmt.Printf("ScanIndent end -------------------\n")
	return ns.Token

}
