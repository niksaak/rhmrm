package asm

import "testing"

type stringWriter struct {
	string
}

func (s *stringWriter) Write(p []byte) (n int, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	s.string += string(p)
	return len(p), nil
}

var ast = &ProgramNode{ // :foo mov r9, r0 ;;;; clear r9
	clauses: []Node{
		&LabelNode{
			name: "foo",
		},
		&InstructionNode{
			op: "mov",
			operands: []Node{
				&RegisterNode{
					kind: generalRegisterKind,
					index: 9,
				},
				&RegisterNode{
					kind: generalRegisterKind,
					index: 0,
				},
			},
		},
		&CommentNode{
			level: 3,
			comment: "clear r9",
		},
	},
}

var astString = `program:(
  label:foo
  instruction:(mov register:r9 register:r0)
  comment:(3 "clear r9")
)
`

func TestPrintAST(t *testing.T) {
	s := &stringWriter{}
	if err := PrintAST(ast, s); err != nil {
		t.Errorf("%v", err)
	}
	if s.string != astString {
		t.Fail()
		t.Log("expected:")
		t.Log(astString)
		t.Log("got:")
		t.Log(s.string)
	}
}
