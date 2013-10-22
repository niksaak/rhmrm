package rhmrm

import "fmt"

// Byte is the unit of memory, which is 16 bit for this machine.
type Byte uint16

// String returns a string for Byte.
func (b Byte) String() string {
	return fmt.Sprintf("%x", uint16(b))
}

// Word is a unit of the machine, 32 bit.
type Word int32

// String returns string representation of a Word.
func (w Word) String() string {
	return fmt.Sprintf("%x", int32(w))
}

// Register is a global register, which is also a top of its' stack
// of 256 Words.
type Register struct {
	stk [256]Word
	top uint8
}

// Val returns the value of a register
func (r Register) Val() Word {
	return r.stk[r.top]
}

// Set method changes the value of a register to n
func (r *Register) Set(n Word) {
	r.stk[r.top] = n
}

// Push copies the value or a register into the stack
func (r *Register) Push() {
	r.stk[r.top+1] = r.stk[r.top]
	r.top++
}

// Pop decrements stack top pointer
func (r *Register) Pop() {
	r.top--
}

// Peek returns value at specified position on the stack
func (r Register) Peek(n Word) Word {
	return r.stk[r.top-(uint8(n)+1)]
}

// String returns string representation or a register
func (r Register) String() string {
	return fmt.Sprint(r.Val())
}

// Clear resets register and rewinds its' stack
func (r *Register) Clear() {
	r.top = 0
	r.Set(Word(0))
}

// Interrupt can be signalled with INT instruction and is returned by
// Step() and Run() methods of the machine
type Interrupt int16
