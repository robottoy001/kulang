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
	"bytes"
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

func (self *NinjaScanner) Reset(ct []byte) {
	self.Content = ct

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
	self.LastToken = tok
	self.Token = tok
	self.Pos.LineNo = 1
	self.Pos.Offset = 0
	self.Pos.LineStart = 0
}

func (self *NinjaScanner) isIdentifier(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '.' || ch == '-' {
		return true
	}
	return false
}

func (self *NinjaScanner) isSimpleIdentifier(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
		return true
	}
	return false
}

func (self *NinjaScanner) skipWhiteSpace() {
	for {
		ch := self.peek()
		switch ch {
		case ' ', '\t':
			self.advance()
			continue
		}
		break
	}
}

func (self *NinjaScanner) newLine() {
	self.Pos.LineNo += 1
	self.Pos.LineStart = self.Pos.Offset
}

func (self *NinjaScanner) GetToken() Token {
	self.Token.Literal = string(self.Content[self.Token.Loc.Start:self.Token.Loc.End])
	return self.Token
}

func (self *NinjaScanner) BackwardToken() {
	self.Token = self.LastToken
	self.Pos.LineNo = self.LastToken.Loc.LineNo
	self.Pos.Offset = self.LastToken.Loc.Start
	self.Pos.LineStart = self.LastToken.Loc.LineStart
}

func (self *NinjaScanner) PeekToken(expected uint8) bool {
	nextToken := self.NextToken()
	if nextToken.Type == expected {
		return true
	}
	self.BackwardToken()
	return false
}

func (self *NinjaScanner) tokenStart() {
	self.Token.Loc.LineNo = self.Pos.LineNo
	self.Token.Loc.Start = self.Pos.Offset
	self.Token.Loc.LineStart = self.Pos.LineStart
}

func (self *NinjaScanner) tokenEnd() {
	self.Token.Loc.End = self.Pos.Offset
	self.LastToken = self.Token
}

func (self *NinjaScanner) NextToken() Token {
	for {
		self.tokenStart()
		ch := self.peek()
		switch {
		case ch == byte(' ') || ch == byte('\t'):
			// from new line or start of file
			if self.LastToken.Type == TokenNewLine {
				self.skipWhiteSpace()
				self.Token.Type = TokenIndent
				self.tokenEnd()
				return self.Token
			}
			self.skipWhiteSpace()
			continue
		case ch == byte('\r'):
			self.advance()
			if self.peek() == byte('\n') {
				self.advance()
			}
			self.newLine()
			self.Token.Type = TokenNewLine
			self.tokenEnd()
			return self.Token
		case ch == byte('\n'):
			self.newLine()
			self.advance()
			self.Token.Type = TokenNewLine
			self.tokenEnd()
			return self.Token
		case ch == byte('#'):
			self.advance()
			self.skipComment()
			continue
		case ch == byte('|'):
			err := self.scanPipe()
			if err != nil {
				log.Fatal(err)
			}
			self.tokenEnd()
			return self.Token
		case ch == ':':
			self.Token.Type = TokenColon
			self.advance()
			self.tokenEnd()
			return self.Token
		case ch == '=':
			self.Token.Type = TokenEquals
			self.advance()
			self.tokenEnd()
			return self.Token
			// keywords
			// build, rule,pool, default
			// include, subninja
		case self.isIdentifier(ch) == true:
			err := self.scanIdentifier()
			if err != nil {
				log.Fatal(err)
			}
			self.tokenEnd()
			return self.Token
		// variable
		case ch == INVALID_CHAR:
			self.Token.Type = TokenEof
			self.tokenEnd()
			return self.Token
		default:
			// error
			log.Panicf("unexpected token (%s) in Line: %d, Col: %d",
				string(ch), self.Pos.LineNo, self.Pos.Offset-self.Pos.LineStart)
			return self.Token
		}
	}
}

// Iterator
func (self *NinjaScanner) advance() {
	self.Pos.Offset += 1
}
func (self *NinjaScanner) backward() {
	self.Pos.Offset -= 1
}

func (self *NinjaScanner) hasNext() bool {
	return self.Pos.Offset <= uint32((len(self.Content) - 1))
}

func (self *NinjaScanner) peek() uint8 {
	if self.hasNext() {
		return self.Content[self.Pos.Offset]
	}
	return INVALID_CHAR
}

func (self *NinjaScanner) next() uint8 {
	n := self.peek()
	self.advance()
	return n
}

func (self *NinjaScanner) skipComment() {
	for {
		ch := self.next()
		switch ch {
		case byte('\r'): // \r
			if self.peek() == byte('\n') {
				self.advance()
			}
			self.newLine()
			return
		case '\n':
			self.newLine()
			return
		}
	} //for
}

func (self *NinjaScanner) scanPipe() error {
	self.advance()

	ch := self.peek()
	switch ch {
	case '@':
		self.Token.Type = TokenPipeAt
		self.advance()
		break
	case '|':
		self.Token.Type = TokenPipe2
		self.advance()
		break
	default:
		self.Token.Type = TokenPipe
	}
	return nil
}

func (self *NinjaScanner) getIdentifierType(identifier string) uint8 {
	if token_type, ok := keyWords[identifier]; ok {
		return token_type
	}

	return TokenIdent
}

func (self *NinjaScanner) scanIdentifier() error {
	start := self.Pos.Offset
Loop:
	for {
		self.advance()
		ch := self.peek()
		switch {
		case self.isIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	self.Token.Type = self.getIdentifierType(string(self.Content[start:self.Pos.Offset]))
	self.Token.Loc.Start = start
	self.Token.Loc.End = self.Pos.Offset
	return nil
}

func (self *NinjaScanner) scanSimpleIdentifier() error {
	start := self.Pos.Offset
Loop:
	for {
		self.advance()
		ch := self.peek()
		switch {
		case self.isSimpleIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	self.Token.Type = self.getIdentifierType(string(self.Content[start:self.Pos.Offset]))
	self.Token.Loc.Start = start
	self.Token.Loc.End = self.Pos.Offset
	return nil
}

func (self *NinjaScanner) ScanVarValue(path bool) (*VarString, error) {
	self.skipWhiteSpace()
	var value VarString
	var tmpStr bytes.Buffer
	var strtype StrType = ORGINAL
	var err error = nil

	tmpStr.Grow(1024)
	self.tokenStart()
Loop:
	for {
		ch := self.peek()
		self.advance()
		switch {
		case ch == '$':
			next := self.peek()
			// escape
			if next == ':' || next == '$' || next == ' ' {
				tmpStr.WriteByte(next)
				self.advance()
				break
			}
			// continue
			if next == '\n' {
				self.advance()
				self.skipWhiteSpace()
				self.newLine()
				break
			}
			if next == '\r' {
				self.advance()
				if self.peek() == '\n' {
					self.advance()
				}
				self.skipWhiteSpace()
				self.newLine()
				break
			}

			// variable name
			if next == '{' {
				// save normal string if have
				if tmpStr.Len() != 0 {
					value.Append(tmpStr.String(), strtype)
				}

				self.advance()

				tmpStr.Truncate(0)
				// scan ${varname}
				strtype = VARIABLE
				err := self.scanIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(string(self.Content[self.Token.Loc.Start:self.Token.Loc.End]), strtype)
			}

			if self.isSimpleIdentifier(next) {
				if tmpStr.Len() != 0 {
					value.Append(tmpStr.String(), strtype)
				}

				tmpStr.Truncate(0)
				// scan ${varname}
				strtype = VARIABLE
				err := self.scanSimpleIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(string(self.Content[self.Token.Loc.Start:self.Token.Loc.End]), strtype)
				strtype = ORGINAL
			}
			break
		case ch == '}':
			// reset to original string
			strtype = ORGINAL
			break
		case ch == '\r' || ch == '\n':
			if path {
				self.backward()
			} else {
				self.newLine()
			}
			self.Token.Type = TokenNewLine
			break Loop
		case ch == ' ' || ch == '|' || ch == ':':
			if path {
				self.backward()
				break Loop
			}
			tmpStr.WriteByte(ch)
		case ch == INVALID_CHAR:
			break Loop
		default:
			tmpStr.WriteByte(ch)
		}
	}
	if tmpStr.Len() != 0 {
		value.Append(tmpStr.String(), strtype)
	}

	self.tokenEnd()

	return &value, err
}

func (self *NinjaScanner) ScanIdent() Token {
	self.skipWhiteSpace()
	start := self.Pos.Offset
	self.tokenStart()
Loop:
	for {
		self.advance()
		ch := self.peek()
		switch {
		case self.isIdentifier(ch):
			continue
		default:
			break Loop
		}
	}

	literal := self.Content[start:self.Pos.Offset]
	self.Token.Type = TokenIdent
	self.Token.Literal = string(literal)
	self.tokenEnd()
	//fmt.Printf("ScanIndent end -------------------\n")
	return self.Token

}
