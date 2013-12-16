package machine

import "testing"

func TestInstructionGetters(t *testing.T) {
	t.Parallel()
	var ti Instruction = 0xc0fe
	if r, want := ti.Op(), Word(ti) & 0x3f; r != want {
		t.Errorf("Got %#x with .op(), want %#x", r, want)
	}
	if r, want := ti.A(), Word(ti) >> 6 & 0x1f; r != want {
		t.Errorf("Got %#x with .a(), want %#x", r, want)
	}
	if r, want := ti.B(), Word(ti) >> 11 & 0x1f; r != want {
		t.Errorf("Got %#x with .b(), want %#x", r, want)
	}
	if r, want := ti.C(), Word(ti) >> 6; r != want {
		t.Errorf("Got %#x with .c(), want %#x", r, want)
	}
} 

func TestInstructionDecouplers(t *testing.T) {
	t.Parallel()
	ti := MkInstruction2(OP_ADD, R+9, R+18)
	if r := ti.Op(); r != OP_ADD {
		t.Errorf("Op is %x, want %x", r, OP_ADD)
	}
	if r := ti.A(); r != R+9 {
		t.Errorf("A is %x, want %x", r, R+9)
	}
	if r := ti.B(); r != R+18 {
		t.Errorf("B is %x, want %x", r, R+18)
	}
}

func TestInstructionNotOverlaps(t *testing.T) {
	t.Parallel()
	ti := MkInstruction2(0x3f, 0, 0x1f)
	op, a, b := ti.Op(), ti.A(), ti.B()

	if op != 0x3f {
		t.Errorf("Op is %x, want %x", op, 0x3f)
	}
	if a != 0 || b != 0x1f {
		t.Errorf("a, b are %x, %x, want %x", a, b, 0, 0x1f)
	}
}
