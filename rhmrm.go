// Rhmrm is a package with RISCy Harward Multistack Register Machine.
package rhmrm

// Registers, 8 general and two special (PC and EX).
const (
	R_0  = iota
	R_1  = iota
	R_2  = iota
	R_3  = iota
	R_4  = iota
	R_5  = iota
	R_6  = iota
	R_7  = iota
	R_PC = iota // Program Counter
	R_EX = iota // EXtra, set by arithmetic, binary and test operations
)

/* Opcodes for, ahem, Rhmrm.

Notes:
    + Arithmetic operators set EX to higher 32 bits of operation result
    + Arithmetic operands all treated as signed
    + On Division By Zero, DIV and MOD operators set x to 0 and signal interrupt
*/
const (
	_ = iota // for future expansion
	// op            args  effect
	OP_MOV = iota // x y   Rx := Ry
	OP_MVZ = iota // x     Rx := 0
	OP_PSH = iota // x     Rx -> stack(x)
	OP_POP = iota // x     Rx <- stack(x)
	OP_PAL = iota // x     stack(x) cleared
	OP_PEK = iota // x y   Rx <- stack(x)[Ry]

	OP_ADD = iota // x y   Rx := Rx + Ry       EX := (Rx + Ry) >> 32
	OP_ADX = iota // x y   Rx := Rx + Ry + EX  EX := (Rx + Ry) >> 32
	OP_SUB = iota // x y   Rx := Rx - Ry       EX := (Rx - Ry) >> 32
	OP_SBX = iota // x y   Rx := Rx - Ry + EX  EX := (Rx - Ry) >> 32
	OP_MUL = iota // x y   Rx := Rx * Ry       EX := (Rx * Ry) >> 32
	OP_DIV = iota // x y   Rx := Rx / Ry       EX := (Rx << 32) / Ry
	OP_MOD = iota // x y   Rx := Rx % Ry

	OP_AND = iota // x y   Rx := Rx & Ry
	OP_BOR = iota // x y   Rx := Rx | Ry
	OP_XOR = iota // x y   Rx := Rx ^ Ry
	OP_SHL = iota // x y   Rx := Rx << Ry      EX := (Rx << Ry) >> 32
	OP_SHR = iota // x y   Rx := Rx >> Ry      EX := (Rx >> Ry) << 32

	OP_XAD = iota // x y   EX := Rx + Ry
	OP_XSB = iota // x y   EX := Rx - Ry
	OP_JPN = iota // x:y   PC := PC + x:y
	OP_JPZ = iota // x:y   if EX == 0 { PC := PC + x:y }
	OP_JPG = iota // x:y   if EX > 0 { PC := PC + x:y }
	OP_JPL = iota // x:y   if EX < 0 { PC := PC + x:y }

	OP_INT = iota // n     send interrupt with message n; PC++
	OP_JMP = iota // nn    PC := [PC+1]:[PC+2];
	OP_LOA = iota // x nn  Rx := [n]; PC += 2
	OP_SET = iota // x nn  Rx := [PC+1]:[PC+2]; PC += 2
)

// Interrupts
const (
	INT_ZERO    Interrupt = iota
	INT_BADOP   Interrupt = -iota
	INT_BADARG  Interrupt = -iota
	INT_ZERODIV Interrupt = -iota
)
