package parser

import (
	"fmt"
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
	self.Pos.LineNo = 0
	self.Pos.Offset = 0
}

func (self *NinjaScanner) GetToken() Token {
	return self.Token
}

func (self *NinjaScanner) isIdentifier(ch byte) bool {
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
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

func (self *NinjaScanner) NextToken() {
	for {
		self.skipWhiteSpace()
		ch := self.peek()
		switch {
		case ch == byte('\r'):
			self.advance()
			if self.peek() == byte('\n') {
				self.advance()
			}
			self.Pos.LineNo += 1
			self.Token.Type = TOKEN_NEWLINE
			continue
		case ch == byte('\n'):
			self.Pos.LineNo += 1
			self.advance()
			self.Token.Type = TOKEN_NEWLINE
			continue
		case ch == byte('#'):
			self.advance()
			self.skipComment()
			continue
		case ch == byte('|'):
			tk, err := self.scanPipe()
			if err != nil {
				panic(err)
			}
			self.Token = tk
			return
		case ch == ':':
			self.Token.Type = TOKEN_COLON
			self.advance()
			return
		case ch == '=':
			self.Token.Type = TOKEN_EQUALS
			self.advance()
			return
			// keywords
			// build, rule,pool, default
			// include, subninja
		case (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_':
			tk, err := self.scanIdentifier()
			if err != nil {
				panic(err)
			}
			self.Token = tk
			return
		// variable
		case ch == INVALID_CHAR:
			self.Token.Type = TOKEN_EOF
			return
		default:
			// error
			errmsg := fmt.Sprintf("unexpected token in Line: %d, Col: %d", self.Pos.LineNo, self.Pos.Offset)
			panic(errmsg)
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
		//ch := self.peek()
		//self.advance()
		ch := self.next()
		switch ch {
		case 0x0D: // \r
			if self.peek() == 0x0A {
				self.advance()
			}
			self.Pos.LineNo += 1
			return
		case 0x0A:
			self.Pos.LineNo += 1
			return
		}
	} //for
}

func (self *NinjaScanner) scanPipe() (Token, error) {
	self.advance()

	var tok Token
	ch := self.peek()
	switch ch {
	case '@':
		tok.Type = TOKEN_PIPEAT
		break
	case '|':
		tok.Type = TOKEN_PIPE2
		break
	default:
		tok.Type = TOKEN_PIPE
	}
	return tok, nil
}

func (self *NinjaScanner) getIdentifierType(identifier string) uint8 {
	if token_type, ok := keyWords[identifier]; ok {
		return token_type
	}

	return TOKEN_IDENT
}

func (self *NinjaScanner) scanIdentifier() (Token, error) {
	var tk Token
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
	tk.Type = self.getIdentifierType(string(literal))
	tk.Literal = string(literal)
	return tk, nil
}

func (self *NinjaScanner) ScanVarValue() (VarString, error) {
	self.skipWhiteSpace()
	var value VarString
	var tmpStr string
	var strtype StrType = ORGINAL
Loop:
	for {
		ch := self.peek()
		self.advance()
		switch {
		case ch == '$':
			next := self.peek()
			// escape
			if next == '$' || next == ' ' {
				tmpStr += string(next)
				break
			}
			// continue
			if next == '\n' {
				self.skipWhiteSpace()
				break
			}
			if next == '\r' {
				self.advance()
				if self.peek() != '\n' {
					self.backward()
				}
				self.skipWhiteSpace()
				break
			}

			// variable name
			if next == '{' {
				// save normal string if have
				if tmpStr != "" {
					value.Append(tmpStr, strtype)
				}
				tmpStr = ""

				// scan ${varname}
				strtype = VARIABLE
				self.advance()
				tok, err := self.scanIdentifier()
				if err != nil {
					panic(err)
				}
				value.Append(tok.Literal, strtype)
			}
			break
		case ch == '}':
			// reset to original string
			strtype = ORGINAL
			break
		case ch == '\r' || ch == '\n':
			break Loop
		case ch == INVALID_CHAR:
			break Loop
		default:
			tmpStr += string(ch)
		}
	}
	return value, nil
}
