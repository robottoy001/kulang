package main

import (
	"fmt"
	"log"
	"testing"
)

func TestSkipComment(t *testing.T) {
	cases := []string{
		"#abcd\r\n|",
		"#abcd\r|",
		"#abcd\n|",
	}

	for _, s := range cases {
		if st := testSkipComment(s); st == PARSER_ERROR {
			t.FailNow()
		}
	}
}

func testSkipComment(s string) State {
	ninja_scanner := &NinjaScanner{
		&Scanner{
			Content: []byte(s),
			Pos: &Position{
				Offset: 0,
				LineNo: 0,
			},
		},
	}

	if !ninja_scanner.PeekToken(TOKEN_PIPE) {
		return PARSER_ERROR
	}
	return PARSER_SUCCESS
}

func TestScanPipe(t *testing.T) {
	cases := map[string]uint8{
		"|":  TOKEN_PIPE,
		"|@": TOKEN_PIPEAT,
		"||": TOKEN_PIPE2,
	}

	for k, v := range cases {
		ninja_scanner := &NinjaScanner{
			&Scanner{
				Content: []byte(k),
				Pos: &Position{
					Offset: 0,
					LineNo: 0,
				},
			},
		}
		ninja_scanner.NextToken()
		token := ninja_scanner.GetToken()
		if token.Type != v {
			t.FailNow()
		}
	}
}

func TestScanIdentifier(t *testing.T) {
	cases := map[string]uint8{
		"build":    TOKEN_BUILD,
		"rule":     TOKEN_RULE,
		"pool":     TOKEN_POOL,
		"default":  TOKEN_DEFUALT,
		"include":  TOKEN_INCLUDE,
		"subninja": TOKEN_SUBNINJA,
	}

	for k, v := range cases {
		ninja_scanner := &NinjaScanner{
			&Scanner{
				Content: []byte(k),
				Pos: &Position{
					Offset: 0,
					LineNo: 0,
				},
			},
		}
		ninja_scanner.NextToken()
		token := ninja_scanner.GetToken()
		if token.Type != v {
			t.FailNow()
		}
	}

}

func TestScanVariable(t *testing.T) {
	cases := map[string]string{
		"bccc = a":             "bccc=a",
		"ba = c$ c":            "ba=c c",
		"ba=c$\r\na":           "ba=ca",
		"ba=c$\r\n\t a":        "ba=ca",
		"ba=c$\ra":             "ba=ca",
		"ba=c$\r a":            "ba=ca",
		"ba=c${foo}":           "ba=cbar",
		"ba=c$$${foo}":         "ba=c$bar",
		"ba=c$$":               "ba=c$",
		"ba=c$$$ \r\n":         "ba=c$ ",
		"ba = cd$ ${foo}d":     "ba=cd bard",
		"ba = cd$\r\n ${foo}d": "ba=cdbard",
	}

	scope := Scope{
		Vars: map[string]string{
			"foo": "bar",
		},
	}

	for k, v := range cases {
		ninja_scanner := &NinjaScanner{
			&Scanner{
				Content: []byte(k),
				Pos: &Position{
					Offset: 0,
					LineNo: 0,
				},
			},
		}

		// var name
		ninja_scanner.NextToken()
		tok := ninja_scanner.GetToken()
		varName := tok.Literal

		// =
		ninja_scanner.NextToken()
		tok = ninja_scanner.GetToken()
		if tok.Type != TOKEN_EQUALS {
			log.Fatal("Expected =, Got ", tok.Literal)
			return
		}

		// value
		valString, err := ninja_scanner.ScanVarValue(false)
		if err != nil {
			t.FailNow()
		}

		result := fmt.Sprintf("%s=%s", varName, valString.Eval(&scope))

		if result != v {
			fmt.Printf("Scan:[%s], expected:[%s], got:[%s]\n", k, v, result)
			t.FailNow()
		}
	}
}
