package rhmrm

import "testing"

func TestInstructionGetters(t *testing.T) {
	t.Parallel()
	var ti Instruction = 0xc0fe
	if r, want := ti.op(), Word(ti) & 0x3f; r != want {
		t.Errorf("Got %#x with .op(), want %#x", r, want)
	}
	if r, want := ti.a(), Word(ti) >> 6 & 0x1f; r != want {
		t.Errorf("Got %#x with .a(), want %#x", r, want)
	}
	if r, want := ti.b(), Word(ti) >> 11 & 0x1f; r != want {
		t.Errorf("Got %#x with .b(), want %#x", r, want)
	}
	if r, want := ti.c(), Word(ti) >> 6; r != want {
		t.Errorf("Got %#x with .c(), want %#x", r, want)
	}
} 

func TestInstructionDecouplers(t *testing.T) {
	t.Parallel()
	ti := MkInstruction2(OP_ADD, R_9, R_18)
	if r := ti.op(); r != OP_ADD {
		t.Errorf("Op is %x, want %x", r, OP_ADD)
	}
	if r := ti.a(); r != R_9 {
		t.Errorf("A is %x, want %x", r, R_9)
	}
	if r := ti.b(); r != R_18 {
		t.Errorf("B is %x, want %x", r, R_18)
	}
}

func TestInstructionNotOverlaps(t *testing.T) {
	t.Parallel()
	ti := MkInstruction2(0x3f, 0, 0x1f)
	op, a, b := ti.op(), ti.a(), ti.b()

	if op != 0x3f {
		t.Errorf("Op is %x, want %x", op, 0x3f)
	}
	if a != 0 || b != 0x1f {
		t.Errorf("a, b are %x, %x, want %x", a, b, 0, 0x1f)
	}
}
