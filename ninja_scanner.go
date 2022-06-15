package main

import (
	"log"
)

const (
	TOKEN_BUILD uint8 = iota
	TOKEN_COLON
	TOKEN_DEFUALT
	TOKEN_EQUALS
	TOKEN_IDENT
	TOKEN_INCLUDE
	TOKEN_SUBNINJA
	TOKEN_INDENT
	TOKEN_NEWLINE
	TOKEN_PIPE
	TOKEN_PIPE2
	TOKEN_PIPEAT
	TOKEN_POOL
	TOKEN_RULE
	TOKEN_EOF
)

var keyWords = map[string]uint8{
	"build":    TOKEN_BUILD,
	"rule":     TOKEN_RULE,
	"include":  TOKEN_INCLUDE,
	"subninja": TOKEN_SUBNINJA,
	"pool":     TOKEN_POOL,
	"default":  TOKEN_DEFUALT,
}

// totally compatiable with Ninja
type NinjaScanner struct {
	*Scanner
}

func (self *NinjaScanner) Reset(ct []byte) {
	self.Content = ct

	tok := Token{
		Type:    TOKEN_EOF,
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
	//fmt.Printf("new line %d\n", self.Pos.LineNo)
	self.Pos.LineNo += 1
	self.Pos.LineStart = self.Pos.Offset
}

func (self *NinjaScanner) GetToken() Token {
	return self.Token
}

func (self *NinjaScanner) BackwardToken() {
	self.Token = self.LastToken
	self.Pos.LineNo = self.LastToken.Loc.LineNo
	self.Pos.Offset = self.LastToken.Loc.Start
	self.Pos.LineStart = self.LastToken.Loc.LineStart

	//	fmt.Printf("backToken: type: %d, LineNO: %d, Offset = %d\n",
	//		self.Pos.LineNo, self.Pos.Offset, self.Pos.LineStart)
}

func (self *NinjaScanner) PeekToken(expected uint8) bool {
	//fmt.Printf("Peektoken --begin-----\n")
	nextToken := self.NextToken()
	if nextToken.Type == expected {
		//		fmt.Printf("Peektoken --end true-----\n")
		return true
	}
	//	fmt.Printf("nextToken: %d, expected: %d\n", nextToken.Type, expected)
	self.BackwardToken()
	return false
}

func (self *NinjaScanner) tokenStart() {
	self.Token.Loc.LineNo = self.Pos.LineNo
	self.Token.Loc.Start = self.Pos.Offset
	self.Token.Loc.LineStart = self.Pos.LineStart

	//	fmt.Printf("tokenStart: Tok: Line: %d-Start %d-LineStart %d\n", self.Pos.LineNo,
	//		self.Token.Loc.Start, self.Token.Loc.LineStart)
}

func (self *NinjaScanner) tokenEnd() {
	self.Token.Loc.End = self.Pos.Offset
	self.LastToken = self.Token

	//	fmt.Printf("tokenEnd: pos.LineNo %d, pos.Offset %d, LastType: %d, last.Loc.Start:%d\n",
	//		self.Pos.LineNo, self.Pos.Offset,
	//		self.LastToken.Type, self.LastToken.Loc.Start)
	//	fmt.Printf("tokenEnd: Tok: Line: %d-Start %d-LineStart %d\n", self.Pos.LineNo,
	//		self.Token.Loc.Start, self.Token.Loc.LineStart)
}

func (self *NinjaScanner) NextToken() Token {
	for {
		self.tokenStart()
		ch := self.peek()
		switch {
		case ch == byte(' ') || ch == byte('\t'):
			// from new line or start of file
			if self.LastToken.Type == TOKEN_NEWLINE {
				self.skipWhiteSpace()
				self.Token.Type = TOKEN_INDENT
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
			//fmt.Printf("nexttoken 2.\n")
			self.newLine()
			self.Token.Type = TOKEN_NEWLINE
			self.tokenEnd()
			return self.Token
		case ch == byte('\n'):
			//fmt.Printf("nexttoken 1.\n")
			self.newLine()
			self.advance()
			self.Token.Type = TOKEN_NEWLINE
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
			self.Token.Type = TOKEN_COLON
			self.advance()
			self.tokenEnd()
			return self.Token
		case ch == '=':
			self.Token.Type = TOKEN_EQUALS
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
			self.Token.Type = TOKEN_EOF
			self.tokenEnd()
			return self.Token
		default:
			// error
			log.Panicf("unexpected token (%s) in Line: %d, Col: %d",
				string(ch), self.Pos.LineNo, self.Pos.Offset-self.Pos.LineStart)
			return self.Token
		}
	}
	return Token{Type: TOKEN_EOF}
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
		//ch := self.peek()
		//self.advance()
		ch := self.next()
		switch ch {
		case byte('\r'): // \r
			if self.peek() == byte('\n') {
				self.advance()
			}
			//fmt.Printf("skipComment.2.\n")
			self.newLine()
			return
		case '\n':
			//fmt.Printf("skipComment.1.\n")
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
		self.Token.Type = TOKEN_PIPEAT
		self.advance()
		break
	case '|':
		self.Token.Type = TOKEN_PIPE2
		self.advance()
		break
	default:
		self.Token.Type = TOKEN_PIPE
	}
	return nil
}

func (self *NinjaScanner) getIdentifierType(identifier string) uint8 {
	if token_type, ok := keyWords[identifier]; ok {
		return token_type
	}

	return TOKEN_IDENT
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

	literal := self.Content[start:self.Pos.Offset]
	self.Token.Type = self.getIdentifierType(string(literal))
	self.Token.Literal = string(literal)
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

	literal := self.Content[start:self.Pos.Offset]
	self.Token.Type = self.getIdentifierType(string(literal))
	self.Token.Literal = string(literal)
	return nil
}

func (self *NinjaScanner) ScanVarValue(path bool) (VarString, error) {
	self.skipWhiteSpace()
	var value VarString
	var tmpStr string
	var strtype StrType = ORGINAL
	var err error = nil

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
				tmpStr += string(next)
				self.advance()
				break
			}
			// continue
			if next == '\n' {
				self.advance()
				self.skipWhiteSpace()
				//fmt.Printf("scanValValue.1.\n")
				self.newLine()
				break
			}
			if next == '\r' {
				self.advance()
				if self.peek() == '\n' {
					self.advance()
				}
				self.skipWhiteSpace()
				//fmt.Printf("scanValValue.2.\n")
				self.newLine()
				break
			}

			// variable name
			if next == '{' {
				// save normal string if have
				if tmpStr != "" {
					value.Append(tmpStr, strtype)
				}

				self.advance()

				tmpStr = ""
				// scan ${varname}
				strtype = VARIABLE
				err := self.scanIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(self.Token.Literal, strtype)
			}

			if self.isSimpleIdentifier(next) {
				if tmpStr != "" {
					value.Append(tmpStr, strtype)
				}

				tmpStr = ""
				// scan ${varname}
				strtype = VARIABLE
				err := self.scanSimpleIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(self.Token.Literal, strtype)
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
			self.Token.Type = TOKEN_NEWLINE
			break Loop
		case ch == ' ' || ch == '|' || ch == ':':
			if path {
				self.backward()
				break Loop
			}
			tmpStr += string(ch)
		case ch == INVALID_CHAR:
			break Loop
		default:
			tmpStr += string(ch)
		}
	}
	if tmpStr != "" {
		value.Append(tmpStr, strtype)
	}

	self.tokenEnd()

	return value, err
}

func (self *NinjaScanner) ScanIdent() Token {
	self.skipWhiteSpace()
	start := self.Pos.Offset
	//fmt.Printf("ScanIndent begin-------------------\n")
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
	self.Token.Type = TOKEN_IDENT
	self.Token.Literal = string(literal)
	self.tokenEnd()
	//fmt.Printf("ScanIndent end -------------------\n")
	return self.Token

}
