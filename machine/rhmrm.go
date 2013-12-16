package machine

import "fmt"

// General register indexing constants. Example usage:
//   R+9 // index of global register #9
//   V+0 // index of value register #0
//   SP  // index of stack pointer register.
const (
	R = 0 // Global register prefix.
	ZR = 0 // ZR (zero) register.
	RA = R+1 // RA (return address) register.
	S = R+2 // S (saved) register prefix, valid indices from 0 to 7.
	T = R+10 // T (temporary) register prefix, 0 to 7.
	V = R+18 // V (value) register prefix, 0 to 3.
	A = R+22 // A (argument) register prefix, 0 to 7.
	FP = R+30 // FP (frame pointer) register.
	SP = R+31 // SP (stack pointer).
)

// Control registers. Example usage:
//   C+4 // index of control register #4.
//   IR  // index of interrupt return register.
const (
	C = 0 // Control register prefix.
	PC = C+0 // PC (program counter) register.
	EX = C+1 // EX (extra or excess) register.
	IA = C+4 // IA (interrupt address) register.
	IM = C+5 // IM (interrupt message) register.
	IR = C+6 // IR (interrupt return) register.
	FL = C+7 // FL (flags) register.
)

// Control register acces modes.
const (
	AM_SET = iota
	AM_AND
	AM_IOR
	AM_XOR
)

// Base operators.
const (
	OP_IMP = 0x00
	OP_MOV = 0x01
	OP_MTC = 0x02
	OP_MFC = 0x03

	OP_STR = 0x04
	OP_PSH = 0x05
	OP_LOA = 0x06
	OP_POP = 0x07
	OP_MOM = 0x08

	OP_SRL = 0x09

	OP_ADD = 0x10
	OP_ADX = 0x11
	OP_SUB = 0x12
	OP_SBX = 0x13
	OP_MUL = 0x14
	OP_MLI = 0x15
	OP_DIV = 0x16
	OP_DVI = 0x17
	OP_MOD = 0x18
	OP_MDI = 0x19
	OP_INC = 0x1a

	OP_AND = 0x20
	OP_IOR = 0x21
	OP_XOR = 0x22
	OP_BIC = 0x23
	OP_SHL = 0x24
	OP_ASR = 0x25
	OP_SHR = 0x26
	OP_ROL = 0x27
	OP_ROR = 0x28

	OP_TST = 0x29
	OP_TEQ = 0x2a
	OP_CMP = 0x2b
	OP_CMN = 0x2c

	OP_JMP = 0x30
	OP_JLT = 0x31
	OP_JLE = 0x32
	OP_JGT = 0x33
	OP_JGE = 0x34
	OP_JEQ = 0x35
	OP_JNE = 0x36

	OP_SWI = 0x3b
	OP_HWI = 0x3c
	OP_IRE = 0x3d
)

// Immediate operand operators.
const (
	IMP_BRK = 0x00
	IMP_MOV = 0x01
	IMP_MTC = 0x02
	IMP_STR = 0x03
	IMP_PSH = 0x04

	IMP_SRL = 0x05

	IMP_ADD = 0x08
	IMP_ADX = 0x09
	IMP_SUB = 0x0a
	IMP_SBX = 0x0b
	IMP_MUL = 0x0c
	IMP_MLI = 0x0d
	IMP_DIV = 0x0e
	IMP_DVI = 0x0f
	IMP_MOD = 0x10
	IMP_MDI = 0x11
	IMP_INC = 0x12

	IMP_AND = 0x13
	IMP_IOR = 0x14
	IMP_XOR = 0x15
	IMP_BIC = 0x16
	IMP_SHL = 0x17
	IMP_ASR = 0x18
	IMP_SHR = 0x19
	IMP_ROL = 0x1a
	IMP_ROR = 0x1b

	IMP_TST = 0x1c
	IMP_TEQ = 0x1d
	IMP_CMP = 0x1e
	IMP_CMN = 0x1f
)


// Word is the machine unit of data.
type Word uint16

// String represents word in hexadecimal integer form as a string.
func (w Word) String() string {
	return fmt.Sprintf("%x", uint16(w))
}

// FlagsRegister is the c7 or fl register of the machine.
type FlagsRegister Word

// I returns true if I flag is set
func (r FlagsRegister) I() bool {
	return r & 0x8000 != 0
}

// SetI sets or clears I flag
func (r *FlagsRegister) SetI(s bool) {
	if s {
		*r |= 0x8000
	} else {
		*r &^= 0x8000
	}
}

// H returns true if H flag is set
func (r FlagsRegister) H() bool {
	return r & 0x4000 != 0
}

// SetH sets or clears H flag
func (r *FlagsRegister) SetH(s bool) {
	if s {
		*r |= 0x4000
	} else {
		*r &^= 0x4000
	}
}

// S returns true if S flag is set
func (r FlagsRegister) S() bool {
	return r & 1 != 0
}

// SetS sets or clears S flag
func (r *FlagsRegister) SetS(s bool) {
	if s {
		*r |= 1
	} else {
		*r &^= 1
	}
}

// Machine is a thing that processes data fed to it according to
// instructions fed to it.
type Machine struct {
	interrupt struct {
		trigger bool // interrupt trigger
		message Word // interrupt message
	}
	regs      [32]Word      // general registers
	ctrl      [8]Word       // control registers
	text      [0x10000]Word // memory
}

// Reset the machine to its' initial state
func (m *Machine) Reset() {
	m.interrupt.trigger = false
	m.interrupt.message = 0
	for i := range m.regs {
		m.regs[i] = 0
	}
	for i := range m.ctrl {
		m.ctrl[i] = 0
	}
	for i := range m.text {
		m.text[i] = 0
	}
}

// R returns a pointer to a global register.
func (m *Machine) R(r Word) *Word {
	if r > 31 {
		panic(fmt.Sprint("Bad register:", r))
	}
	return &m.regs[r]
}

// C returns a pointer to a control register.
func (m *Machine) C(c Word) *Word {
	if c > 7 {
		panic(fmt.Sprint("Bad control register:", c))
	}
	return &m.ctrl[c]
}

// PC returns program counter of the machine
func (m *Machine) PC() *Word {
	return &m.ctrl[PC]
}

// Mem returns a pointer to a word in Machine memory.
func (m *Machine) Mem(i Word) *Word {
	return &m.text[i]
}

// Text is like Mem(), but points to instruction.
func (m *Machine) Text(i Word) *Instruction {
	return (*Instruction)(&m.text[i])
}

// HWInterrupt triggers hardware interrupt with message i.
func (m *Machine) HWInterrupt(i Word) {
	m.interrupt.trigger = true
	m.interrupt.message = i
}

// Load copies words from slice to Machine memory.
func (m *Machine) Load(text []Word) {
	for i, w := range text {
		if i > 0xffff {
			return
		}
		m.text[i] = w
	}
}

// Step executes one instruction and increments the Program Counter.
func (m *Machine) Step() (interrupt Word, trigger bool) {
	o, args := (*m.Text(*m.PC())).decouple()
	op := op_funcs[o]
	if op == nil {
		return 0xffff, true
	}
	*m.R(0) = 0
	*m.PC()++
	op(m, args...)
	if m.interrupt.trigger {
		m.interrupt.trigger = false
		return m.interrupt.message, true
	} else {
		return 0, false
	}
}
