package parser

import (
	"testing"
)

func TestSkipComment(t *testing.T) {
	ninja_scanner := &NinjaScanner{
		&Scanner{
			Content: []byte("#abcd\r\n|"),
		},
	}

	ninja_scanner.NextToken()
	token := ninja_scanner.GetToken()
	if token.Type != TOKEN_PIPE {
		t.FailNow()
	}
}
