package rhmrm

import "testing"
import "math/rand"
import "time"

func mkran_machine() *Machine {
	rand.Seed(time.Now().UnixNano())
	m := MkMachine()
	for i := range m.reg {
		for j := rand.Intn(255); j > 0; j-- {
			m.Register(uint(i)).Set(Word(j))
			m.Register(uint(i)).Push()
		}
	}
	m.SetPC(uint32(rand.Intn(0x10000)))
	m.SetEX(Word(rand.Intn(0x10000)))
	return m
}

func TestMkMachine(t *testing.T) {
	t.Parallel()
	m := MkMachine()
	if v := m.Register(3).Val(); v != 0 {
		t.Errorf("r3 == %v, want 0", v)
	}
	if v := m.PC(); v != 0 {
		t.Errorf("pc == %v, want 0", v)
	}
	if v := m.EX(); v != 0 {
		t.Errorf("pc == %v, want 0", v)
	}
	if ln, sz := len(m.text), cap(m.text); ln != 0 && sz != 0x10000 {
		t.Logf("len(text) == %v, cap(text) == %v", ln, sz)
		t.Logf("want len(text) == %v, cap(text) == %v", 0, 0x10000)
		t.Fail()
	}
}

func TestMachineReset(t *testing.T) {
	t.Parallel()
	m := mkran_machine()
	m.Reset()

	r3v, r3stk := m.Register(3).Val(), m.Register(3).top
	if r3v != 0 || r3stk != 0 {
		t.Logf("r3 == %v, stack(3) == %v", r3v, r3stk)
		t.Logf("want r3 == 0, stack(3) == 0")
		t.Fail()
	}
	if pc := m.PC(); pc != 0 {
		t.Errorf("pc == %v, want 0", pc)
	}
	if ex := m.EX(); ex != 0 {
		t.Errorf("ex == %v, want 0", ex)
	}
}

func TestMachineRegister(t *testing.T) {
	t.Parallel()
	m := mkran_machine()

	if m.Register(R_0) != &m.reg[R_0] {
		t.Error("Register() returns wrong value")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered: ", r)
		}
	}()
	m.Register(9001)
}

func TestMachineSpecials(t *testing.T) {
	t.Parallel()
	m := mkran_machine()

	if m.PC() != m.pc {
		t.Error("PC() does not return pc. Did you do something funny?")
	}
	if m.EX() != m.ex {
		t.Error("EX() does not return ex. Did you do something funny?")
	}
}


func TestMachine(t *testing.T) {
	t.Parallel()
	i1 := MkInstruction(OP_SET, R_1, 0x10c)
}
