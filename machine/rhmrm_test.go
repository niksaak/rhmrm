package machine

import "testing"

// convenient abbreviations
var i1, i2 = WMkInstruction1, WMkInstruction2

// mk_machine creates machine with superwisor mode bit set.
func mk_machine() *Machine {
	m := new(Machine)
	(*FlagsRegister)(m.C(FL)).SetS(true)
	return m
}

// exec_until_interrupt calls m.Step until the first interrupt,
// but no more than step_max times.
func exec_until_interrupt(m *Machine, step_max int) (msg Word, t bool) {
	for i := 0; !t && i < step_max; i++ {
		msg, t = m.Step()
	}
	return msg, t
}

func TestSimpleAddition(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	*m.R(1), *m.R(2) = 1, 2
	text := []Word{
		WMkInstruction2(OP_ADD, R+1, R+2),
		WMkInstruction1(OP_HWI, 9),
	}
	m.Load(text)

	for i := Word(0); i != 9; {
		instr := *m.Text(*m.PC())
		t.Logf("r1: %04x; r2: %04x; pc %04x",
			*m.R(1), *m.R(2), *m.PC())
		t.Log(instr)
		i, _ = m.Step()
	}
	if r := *m.R(R+1); r != 3 {
		t.Errorf("r1 == %x, want %x", r, 3)
	}
}

func TestImmediateLoadingAddition(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	text := []Word{
		WMkInstruction2(OP_IMP, IMP_MOV, R+1),
		2,
		WMkInstruction2(OP_IMP, IMP_MOV, R+2),
		1,
		WMkInstruction2(OP_ADD, R+1, R+2),
		WMkInstruction1(OP_HWI, 9),
		WMkInstruction1(OP_JMP, -1),
	}
	m.Load(text)
	for i := false; i != true; {
		instr := Instruction(*m.Mem(*m.PC()))
		t.Logf("r1: %04v; r2: %04v; pc: %04v",
			*m.R(1), *m.R(2), *m.C(PC))
		t.Logf("%2d %2d, %2d", instr.Op(), instr.A(), instr.B())
		_, i = m.Step()
	}
	if r := *m.R(R+1); r != 3 {
		t.Errorf("r1 == %x, want %x", r, 3)
	}
}

/* Fibonacci function:
:fib    mov v0, zr
    imp mov v1, 1
        cmp a0, zr
        jeq _ret
:_loop  mov t0, v0
        add t0, v1
        mov v0, v1
        mov v1, t0
        inc a0, -1
        cmp a0, zr
        jgt _loop
:_ret   srl rz, ra
*/

func TestFibFunction(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	i1, i2 := WMkInstruction1, WMkInstruction2
	text := []Word{ // handmaid assembly ftw
		i2(OP_IMP, IMP_MOV, A+0),
		9,
		i2(OP_IMP, IMP_SRL, RA),
		5,
		i1(OP_HWI, 9),
		i2(OP_MOV, V+0, ZR), // :fib
		i2(OP_IMP, IMP_MOV, V+1),
		1,
		i2(OP_CMP, A+0, ZR),
		i1(OP_JEQ, 8),
		i2(OP_MOV, T+0, V+0), // :_loop
		i2(OP_ADD, T+0, V+1),
		i2(OP_MOV, V+0, V+1),
		i2(OP_MOV, V+1, T+0),
		i2(OP_INC, A+0, -1),
		i2(OP_CMP, A+0, ZR),
		i1(OP_JGT, -7),
		i2(OP_SRL, ZR, RA), // :_ret
	}
	m.Load(text)
	for i, j := false, 0; !i && j < 80; j++ {
		instr := *m.Text(*m.PC())
		t.Logf("a0: %04v; v0: %04v; ex: %04v, pc: %04v",
			*m.R(A+0), *m.R(V+0), *m.C(EX), *m.PC())
		t.Log(instr)
		_, i = m.Step()
	}
	if ret := *m.R(V+0); ret != 34 {
		t.Errorf("fib(9) returns %v (%d), want %d", ret, ret, 34)
	}
}

func TestOP_SRL(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	text := []Word{
		i2(OP_SRL, R+0, R+1),
		9: i1(OP_HWI, 9),
	}
	m.Load(text)
	*m.R(R+1) = 9
	msg, _ := exec_until_interrupt(m, 2)
	if msg != 9 {
		t.Errorf("msg=%v, want %v", msg, Word(9))
	}
}

func TestIMP_SRL(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	text := []Word{
		i2(OP_IMP, IMP_SRL, R+0),
		9,
		9: i1(OP_HWI, 9),
	}
	m.Load(text)
	msg, _ := exec_until_interrupt(m, 3)
	if msg != 9 {
		t.Errorf("msg=%v, want %v", msg, Word(9))
	}
}
