package rhmrm

import "testing"

func TestOperatorString(t *testing.T) {
	t.Parallel()
	if s := Operator(0).String(); s != "?0?" {
		t.Errorf("Operator(0).String() is %s, want \"???\"", s)
	}
	if s := Operator(OP_MOV).String(); s != "mov" {
		t.Errorf("Operator(OP_MOV).String() is %s, want \"mov\"", s)
	}
	if s := Operator(OP_JPN).String(); s != "jpn" {
		t.Errorf("Operator(OP_JPN).String() is %s, want \"jpn\"", s)
	}
	if s := Operator(9001).String(); s != "?2329?" {
		t.Errorf("Operator(9001).String() is %s, want \"???\"", s)
	}
}

func TestOperandString(t *testing.T) {
	t.Parallel()
	if s := Operand(R_3).String(); s != "r3" {
		t.Errorf("Operand(R_3).String() is %s, want \"r3\"", s)
	}
	if s := Operand(9001).String(); s != "2329" {
		t.Errorf("Operand(9001).String() is %s, want \"r3\"", s)
	}
}

func TestInstructionString(t *testing.T) {
	ins := MkInstruction(OP_SET, R_3, R_0)
	tgt := "set r3, r0"
	if s := ins.String(); s != tgt {
		t.Errorf("Instruction string is %s, want %s", s, tgt)
	}
}

func TestMkInstruction(t *testing.T) {
	t.Parallel()
	if i := MkInstruction(OP_MOV, R_0, R_0); i[0] != 1 {
		t.Errorf("instruction(mov, r0, r0) => %x, want 1", i)
	}
}

func TestInstructionGetters(t *testing.T) {
	t.Parallel()
	i := MkInstruction(OP_MOV, R_3, R_1)
	if i.X() != R_3 {
		t.Errorf("i.X() => %v, not R_3", i.X())
	}
	if i.Y() != R_1 {
		t.Errorf("i.Y() => %v, not R_1", i.Y())
	}
	if i.XY() != Operand(i[0] >> 6) {
		t.Errorf("i.XY() => %v, not %x", i.XY(), R_1 | (R_3 << 5))
	}
}
