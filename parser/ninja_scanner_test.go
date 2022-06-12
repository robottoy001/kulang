package parser

import (
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
		},
	}

	ninja_scanner.NextToken()
	token := ninja_scanner.GetToken()
	if token.Type != TOKEN_PIPE {
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
			},
		}
		ninja_scanner.NextToken()
		token := ninja_scanner.GetToken()
		if token.Type != v {
			t.FailNow()
		}
	}

}
