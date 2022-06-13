package main

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
	Reset([]byte)
	ScanVarValue() (VarString, error)
}

func NewScanner(scanner_type uint8, content []byte) ScannerI {
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
