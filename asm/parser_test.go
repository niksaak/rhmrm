package asm

import "testing"

// mkErrFunction returns an ErrorHandler which reports every error
// to test runner.
func mkErrFunction(t *testing.T) ErrorHandler {
	return func(p Position, msg string) {
		t.Errorf("%v: %s\n", &p, msg)
	}
}

var asm = `
:foo mov r9, r0    ;; full clause
     add r1, 1eech ;; no label, suffixed hex number
:_bar              ;; no instruction, private label
     mtc &pc, r1   ;; prefixed control register
`

func TestParseProgram(t *testing.T) {
	err := mkErrFunction(t)
	l := new(Lexer).Init([]byte(asm), "", err)
	p := new(Parser).Init(l)
	program := p.ParseProgram()
	// TODO: check resulting ast in a proper way.
	if program == nil {
		t.Errorf("program was not parsed")
	}
	if n := p.ErrorCount; n > 0 {
		t.Errorf("got %d parse errors", n)
	}
}
