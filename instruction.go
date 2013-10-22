package rhmrm

import "fmt"

// Operator part of the instruction.
type Operator uint16

var opstrings []string = []string{
	OP_MOV: "mov",
	OP_MVZ: "mvz",
	OP_PSH: "psh",
	OP_POP: "pop",
	OP_PAL: "pal",
	OP_PEK: "pek",
	OP_ADD: "add",
	OP_ADX: "adx",
	OP_SUB: "sub",
	OP_SBX: "sbx",
	OP_MUL: "mul",
	OP_DIV: "div",
	OP_MOD: "mod",
	OP_AND: "and",
	OP_BOR: "bor",
	OP_XOR: "xor",
	OP_SHL: "shl",
	OP_SHR: "shr",
	OP_XAD: "xad",
	OP_XSB: "xsb",
	OP_JPN: "jpn",
	OP_JPZ: "jpz",
	OP_JPG: "jpg",
	OP_JPL: "jpl",
	OP_INT: "int",
	OP_JMP: "jmp",
	OP_LOA: "loa",
	OP_SET: "set",
}

// String returns string representation of an opcode.
func (o Operator) String() string {
	n := uint16(o)
	if n > 0 && int(n) < len(opstrings) {
		return opstrings[n]
	}
	return fmt.Sprintf("?%x?", uint16(n))
}

// Operand part of the instruction.
type Operand uint16

// Registerp returns true if operand is a global register
func (o Operand) Registerp() bool {
	if uint16(o) < R_7 {
		return true
	} else {
		return false
	}
}

var regstrings []string = []string{
	R_0:  "r0",
	R_1:  "r1",
	R_2:  "r2",
	R_3:  "r3",
	R_4:  "r4",
	R_5:  "r5",
	R_6:  "r6",
	R_7:  "r7",
	R_PC: "pc",
	R_EX: "ex",
}

// String returns string representation of an operand.
func (o Operand) String() string {
	if o <= R_EX {
		return regstrings[int(o)]
	} else {
		return fmt.Sprintf("%x", uint(o))
	}
}

// Instruction.
type Instruction []Byte

// instruction does necessary shifts to glue operator and operands together
func instruction(op uint16, x, y uint32) Byte {
	rop := op & 0x3f
	rx := (uint16(x) & 0x1f) << 6
	ry := (uint16(y) & 0x1f) << 11
	return Byte(rop | rx | ry)
}

// MkInstruction creates instruction code and returns it as a slice of Bytes
func MkInstruction(op uint16, args ...uint32) Instruction {
	if op > OP_SET {
		panic(fmt.Sprintf("Bad operator: %v", op))
	}
	ins := make([]Byte, 0, 3)

	switch op {
	case OP_PSH, OP_POP:
		x := args[0] & 0x1f
		ins[0] = instruction(op, x, 0)
	case OP_JPN, OP_JPZ, OP_JPG, OP_JPL:
		x := args[0] & 0x1f
		y := args[0]>>6 & 0x1f
		ins[0] = instruction(op, x, y)
	case OP_INT:
		ins[0] = instruction(op, 0, 0)
		ins[1] = Byte(args[0])
	case OP_JMP:
		n := args[0]
		ins[0] = instruction(op, 0, 0)
		ins[1] = Byte(n & 0xffff)
		ins[2] = Byte(n >> 16)
	case OP_LOA, OP_SET:
		x := args[0] & 0x1f
		n := args[1]
		ins[0] = instruction(op, x, 0)
		ins[1] = Byte(n & 0xffff)
		ins[2] = Byte(n >> 16)
	default:
		ins[0] = instruction(op, args[0], args[1])
	}

	return ins
}

/*
// MkInstruction creates new instruction
func MkInstruction(op uint16, args ...uint32) (instr Instruction,
	word uint32, count uint) {
	if op > OP_SET {
		panic(fmt.Sprint("Bad operator:", op))
	}

	switch op {
	// PSH and POP have only one operand
	case OP_PSH, OP_POP:
		x := args[0] & 0x1f
		return instruction(op, x, 0), 0, 0
	// Short and conditional jumps accept one literal concatenated operand
	case OP_JPN, OP_JPZ, OP_JPG, OP_JPL:
		x := args[0] & 0x1f
		y := (args[0] & (0x1f << 6)) >> 6
		return instruction(op, x, y), 0, 0
	// Interrupts take one word at [PC+1]
	case OP_INT:
		n := args[0]
		return instruction(op, 0, 0), n, 1
	// Longjump takes two words at [PC+2]:[PC+1]
	case OP_JMP:
		n := args[0]
		return instruction(op, 0, 0), n, 2
	// Load instruction takes one operand and [PC+2]:[PC+1]
	case OP_LOA:
		x := args[0] & 0x1f
		n := args[1]
		return instruction(op, x, 0), n, 2
	// Set instruction takes double word [PC+2]:[PC+1]
	case OP_SET:
		x := args[0] & 0x1f
		n := args[1]
		return instruction(op, x, 0), n, 2
	// Other instructions take two operands
	default:
		return instruction(op, args[0], args[1]), 0, 0
	}
}
*/

// String returns a string representation of an instruction.
func (i Instruction) String() string {
	op := i.Op()
	x := i.X()
	y := i.Y()
	return fmt.Sprintf("%v %v, %v", op, x, y)
}

// Op returns the operator of an instruction.
func (i Instruction) Op() Operator {
	n := uint16(i[0])
	return Operator(n & 0x3f)
}

// X returns the first operand of an instruction.
func (i Instruction) X() Operand {
	n := uint16(i[0])
	return Operand(n >> 6 & 0x1f)
}

// Y returns the second operand of an instruction.
func (i Instruction) Y() Operand {
	n := uint16(i[0])
	return Operand(n >> 11 & 0x1f)
}

// XY returns x and y operands concatenated
func (i Instruction) XY() Operand {
	n := uint16(i[0])
	return Operand(n >> 6)
}
