package parser

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

const (
	SCANNER_NINJA uint8 = iota
)

const (
	INVALID_CHAR = 255
)

type Token struct {
	Type    uint8
	Literal string
}

type Position struct {
	Offset uint32
	LineNo uint32
}

type Scanner struct {
	Content []uint8
	Token   Token
	Pos     Position
}

type ScannerI interface {
	NextToken()
	GetToken() Token
}

func NewScanner(scanner_type uint8, content []uint8) ScannerI {
	switch scanner_type {
	case SCANNER_NINJA:
		return &NinjaScanner{&Scanner{
			Content: content,
			Pos: Position{
				Offset: 0,
				LineNo: 0,
			},
		},
		}
	default:
		return nil
	}
	return nil
}
