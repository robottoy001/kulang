package parser

import "fmt"

// totally compatiable with Ninja
type NinjaScanner struct {
	*Scanner
}

func (self *NinjaScanner) GetToken() Token {
	return self.Token
}

func (self *NinjaScanner) NextToken() {
	for {
		ch := self.peek()
		switch ch {
		case byte('#'):
			self.advance()
			self.skipComment()
			continue
		case byte('|'):
			tk, err := self.scanPipe()
			if err != nil {
				panic(err)
			}
			self.Token = tk
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

// scanner case
func (self *NinjaScanner) skipComment() {
	for {
		ch := self.peek()
		switch ch {
		case 0x0D: // \r
			self.advance()
			if self.peek() != 0x0A {
				self.backward()
			} else {
				self.advance()
			}
			self.Pos.LineNo += 1
			return
		case 0x0A:
			self.Pos.LineNo += 1
			self.advance()
			return
		default:
			self.advance()
		}
	} //for
}

func (self *NinjaScanner) scanPipe() (Token, error) {
	tk := Token{
		Type: TOKEN_PIPE,
	}
	return tk, nil
}
