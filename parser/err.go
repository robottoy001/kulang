package parser

type ScannerErr struct {
	msg string
}

func (self *ScannerErr) String() string {
	return self.msg
}
