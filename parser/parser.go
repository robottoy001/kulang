package parser

import (
	"io/ioutil"
	"os"
)

type Scope struct {
	Vaiables map[string]string
	Parent   *Scope
}

type Parser struct {
	Scanner ScannerI
	Scope   *Scope
	// node,edges
}

type ParserI interface {
	Parse() error
	Load(fileName string) error
}

// default: Ninja parser
func NewParser(scope *Scope) ParserI {
	return &NinjaParser{
		&Parser{
			Scanner: NewScanner(SCANNER_NINJA, []byte{}),
			Scope:   scope,
		},
	}
}

func NewParserWithScanner(scanner_type uint8) ParserI {
	return &NinjaParser{
		&Parser{Scanner: NewScanner(scanner_type, []byte{})},
	}
}

func (self *Parser) Load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	ct, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	self.Scanner.Reset(ct)
	return nil
}
