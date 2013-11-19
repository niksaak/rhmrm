package rhmrm

import "testing"

// convenient abbreviations
var i1, i2 = WMkInstruction1, WMkInstruction2

// mk_machine creates machine with superwisor mode bit set.
func mk_machine() *Machine {
	m := new(Machine)
	(*FlagsRegister)(m.C(C_FL)).SetS(true)
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
		WMkInstruction2(OP_ADD, R_1, R_2),
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
	if r := *m.R(R_1); r != 3 {
		t.Errorf("r1 == %x, want %x", r, 3)
	}
}

func TestImmediateLoadingAddition(t *testing.T) {
	t.Parallel()
	m := mk_machine()
	text := []Word{
		WMkInstruction2(OP_IMP, IMP_MOV, R_1),
		2,
		WMkInstruction2(OP_IMP, IMP_MOV, R_2),
		1,
		WMkInstruction2(OP_ADD, R_1, R_2),
		WMkInstruction1(OP_HWI, 9),
		WMkInstruction1(OP_JMP, -1),
	}
	m.Load(text)
	for i := false; i != true; {
		instr := Instruction(*m.Mem(*m.PC()))
		t.Logf("r1: %04v; r2: %04v; pc: %04v",
			*m.R(1), *m.R(2), *m.C(C_PC))
		t.Logf("%2d %2d, %2d", instr.op(), instr.a(), instr.b())
		_, i = m.Step()
	}
	if r := *m.R(R_1); r != 3 {
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
		i2(OP_IMP, IMP_MOV, R_A0),
		9,
		i2(OP_IMP, IMP_SRL, R_RA),
		5,
		i1(OP_HWI, 9),
		i2(OP_MOV, R_V0, R_ZR), // :fib
		i2(OP_IMP, IMP_MOV, R_V1),
		1,
		i2(OP_CMP, R_A0, R_ZR),
		i1(OP_JEQ, 8),
		i2(OP_MOV, R_T0, R_V0), // :_loop
		i2(OP_ADD, R_T0, R_V1),
		i2(OP_MOV, R_V0, R_V1),
		i2(OP_MOV, R_V1, R_T0),
		i2(OP_INC, R_A0, -1),
		i2(OP_CMP, R_A0, R_ZR),
		i1(OP_JGT, -7),
		i2(OP_SRL, R_ZR, R_RA), // :_ret
	}
	m.Load(text)
	for i, j := false, 0; !i && j < 80; j++ {
		instr := *m.Text(*m.PC())
		t.Logf("a0: %04v; v0: %04v; ex: %04v, pc: %04v",
			*m.R(R_A0), *m.R(R_V0), *m.C(C_EX), *m.PC())
		t.Log(instr)
		_, i = m.Step()
	}
	if ret := *m.R(R_V0); ret != 34 {
		t.Errorf("fib(9) returns %v (%d), want %d", ret, ret, 34)
	}
}
