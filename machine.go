package rhmrm

import "fmt"

// Machine struct contains general and special registers, as well
// as executable memory.
type Machine struct {
	reg  [8]Register // general registers
	pc   uint32      // program counter
	ex   Word        // extra
	text []Byte      // program
}

// MkMachine creates new machine
func MkMachine() *Machine {
	m := new(Machine)
	m.text = make([]Byte, 0, 0x10000)
	return m
}

// Reset takes the machine to its' initial state
func (m *Machine) Reset() {
	for i := range m.reg {
		m.reg[i].Clear()
	}
	m.pc = 0
	m.ex = 0
	for i := range m.text {
		m.text[i] = 0
	}
}

// Register method returns the global register of the machine
func (m *Machine) Register(n uint) *Register {
	if n > R_7 {
		panic(fmt.Sprint("Bad register:", n))
	}
	return &m.reg[n]
}

// R is the same as Register() but accepts Operands and doesn't panic
func (m *Machine) R(o Operand) *Register {
	if o > R_7 {
		return nil
	}
	return &m.reg[o]
}

// PC returns value of the machine Program Counter
func (m Machine) PC() uint32 {
	return m.pc
}

// SetPC sets the PC register
func (m *Machine) SetPC(n uint32) {
	m.pc = n
}

func (m *Machine) AddPC(n int32) {
	if n < 0 {
		m.pc -= uint32(-n)
	} else {
		m.pc += uint32(n)
	}
}

// EX returns the value of EXtra register
func (m Machine) EX() Word {
	return m.ex
}

// SetEX sets the EX register
func (m *Machine) SetEX(n Word) {
	m.ex = n
}

// Text returns the byte at index n in text slice
func (m Machine) Text(n uint32) Byte {
	return m.text[n]
}

// FullText returns the text slice
func (m Machine) FullText() []Byte {
	return m.text
}

// LoadText loads text into Machine.
// Text that was there before will be overwritten.
// Does panic on buffer overflow.
func (m *Machine) LoadText(text []Byte) {
	for i, v := range text {
		m.text[i] = v
	}
}

// Step method executes one instruction.
// If any interrupt is triggered during step, it is returned
// as first value, with second value, ok, set to true.
func (m *Machine) Step() (i Interrupt, ok bool) {
	ci := m.Text(m.PC())
	op := int((ci.Op()))

	if op >= len(OpFuncs) {
		return INT_BADOP, true
	}
	i, ok = OpFuncs[op](ci, m)
	m.AddPC(1)

	return i, ok
}
