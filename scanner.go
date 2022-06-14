package main

const (
	SCANNER_NINJA uint8 = iota
)

const (
	INVALID_CHAR = 255
)

type Location struct {
	LineNo    uint32
	Start     uint32
	End       uint32
	LineStart uint32
}

type Token struct {
	Type    uint8
	Literal string
	Loc     Location
}

type Position struct {
	Offset    uint32
	LineNo    uint32
	LineStart uint32
}

func (self *Position) NewLine() {
	self.LineNo += 1
	self.LineStart = self.Offset
}

type Scanner struct {
	Content   []uint8
	Token     Token
	LastToken Token
	Pos       Position
}

type ScannerI interface {
	NextToken() Token
	PeekToken(uint8) bool
	GetToken() Token
	Reset([]byte)
	ScanVarValue(bool) (VarString, error)
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
