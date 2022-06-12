package parser

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
	case tok.Type == TOKEN_IDENT:
		self.parseVariable()
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

func (self *NinjaParser) parseVariable() {
	tok := self.Scanner.GetToken()

	varName := tok.Literal
	// expectd '='
	self.Scanner.NextToken()
	tok = self.Scanner.GetToken()
	if tok.Type != TOKEN_EQUALS {
		log.Fatal("Expected =, Got ", tok.Literal)
		return
	}

	self.parseVarValue()

	//fmt.Printf("Var: %s = %s\n", varName, tok.Literal)

	//self.Scanner.NextToken()
	//tok = self.Scanner.GetToken()
	//if tok.Type != TOKEN_NEWLINE {
	//	log.Fatal("Expected new line, Got ", tok.Literal)
	//}
}

func (self *NinjaParser) parseVarValue() (string, error) {
	valString, err := self.Scanner.ScanVarValue()
	if err != nil {
		panic(err)
	}
	return valString.Eval(self.Scope), err
}
