package rhmrm

type OpFunc func(Instruction, *Machine) (Interrupt, bool)

/* Data ops */

func op_mov(i Instruction, m *Machine) (Interrupt, bool) {
	x, y := m.R(i.X()), m.R(i.Y())
	if x == nil || y == nil {
		return INT_BADARG, true
	}
	x.Set(y.Val())
	return INT_ZERO, false
}

func op_mvz(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	if x == nil {
		return INT_BADARG, true
	}
	x.Set(Word(0))
	return INT_ZERO, false
}

func op_psh(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	if x == nil {
		return INT_BADARG, true
	}
	x.Push()
	return INT_ZERO, false
}

func op_pop(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	if x == nil {
		return INT_BADARG, true
	}
	x.Pop()
	return INT_ZERO, false
}

func op_pal(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	if x == nil {
		return INT_BADARG, true
	}
	val := x.Val()
	x.Clear()
	x.Set(val)
	return INT_BADARG, false
}

func op_pek(i Instruction, m *Machine) (Interrupt, bool) {
	x, y := m.R(i.X()), m.R(i.Y())
	if x == nil || y == nil {
		return INT_BADARG, true
	}
	x.Set(x.Peek(y.Val()))
	return INT_ZERO, false
}

/* Arithmetic ops */

func math_op(x *Register, y *Register, f func(int64, int64) int64) (Word, Word) {
	vx, vy := int64(x.Val()), int64(y.Val())
	r := f(vx, vy)
	val := Word(r)
	ex := Word(r >> 32)
	return val, ex
}

func mkmath_op(f func(x, y int64) int64, plus_ex, zerodiv bool) OpFunc {
	return func(i Instruction, m *Machine) (Interrupt, bool) {
		x, y := m.R(i.X()), m.R(i.Y())
		if x == nil || y == nil {
			return INT_BADARG, true
		}
		if zerodiv && y.Val() == 0 {
			x.Set(Word(0))
			m.SetEX(Word(0))
			return INT_ZERODIV, true
		}
		v, ex := math_op(x, y, f)
		if plus_ex {
			x.Set(v + ex)
		} else {
			x.Set(v)
		}
		m.SetEX(ex)
		return INT_ZERO, false
	}
}

var op_add = mkmath_op(func(x, y int64) int64 { return x + y }, false, false)
var op_adx = mkmath_op(func(x, y int64) int64 { return x + y }, true, false)
var op_sub = mkmath_op(func(x, y int64) int64 { return x - y }, false, false)
var op_sbx = mkmath_op(func(x, y int64) int64 { return x - y }, true, false)
var op_mul = mkmath_op(func(x, y int64) int64 { return x * y }, false, false)

var op_div = mkmath_op(func(x, y int64) int64 { return x / y }, false, true)
var op_mod = mkmath_op(func(x, y int64) int64 { return x % y }, false, true)

var op_and = mkmath_op(func(x, y int64) int64 { return x & y }, false, false)
var op_bor = mkmath_op(func(x, y int64) int64 { return x | y }, false, false)
var op_xor = mkmath_op(func(x, y int64) int64 { return x ^ y }, false, false)

var op_shl = mkmath_op(func(x, y int64) int64 {
	return x << uint(y)
}, false, false)

var op_shr = mkmath_op(func(x, y int64) int64 {
	// TODO: check it this actually works
	rx := x >> uint(y)
	ry := int64(0)
	if y < 32 {
		ry = x << uint(32+(32-y))
	}
	return rx & ry
}, false, false)

/* Control Flow ops */

/*
func int10_to_int16_(n uint16) int16 {
	r := int16(0)
	v := n & 0x200;
	if v == 0 {
		return r
	}
	r = int16(n & (v ^ 0x200))
	v <<= 6
	return r & int16(v)
}
*/

func int10_to_int16(n uint16) int16 {
	if n == 0 {
		return 0
	}
	if n&0x200 != 0 { // if sign bit is set
		return int16(n | 0xfc00) // fill the rest with ones
	}
	return int16(n & 0x3ff) // discard garbage bits
}

func mkctrl_op(test func(int32) bool) OpFunc {
	return func(i Instruction, m *Machine) (Interrupt, bool) {
		xy := uint16(i.XY())
		if test != nil {
			if test(int32(m.EX())) {
				m.AddPC(int32(int10_to_int16(xy)))
			}
		} else {
			m.AddPC(int32(int10_to_int16(xy)))
		}
		return INT_ZERO, false
	}
}

func op_xad(i Instruction, m *Machine) (Interrupt, bool) {
	x, y := m.R(i.X()), m.R(i.Y())
	if x == nil || y == nil {
		return INT_BADARG, false
	}
	_, ex := math_op(x, y, func(m, n int64) int64 { return m + n })
	m.SetEX(ex)
	return INT_ZERO, false
}

func op_xsb(i Instruction, m *Machine) (Interrupt, bool) {
	x, y := m.R(i.X()), m.R(i.Y())
	if x == nil || y == nil {
		return INT_BADARG, true
	}
	_, ex := math_op(x, y, func(m, n int64) int64 { return m - n })
	m.SetEX(ex)
	return INT_ZERO, false
}

var op_jpn = mkctrl_op(nil)
var op_jpz = mkctrl_op(func(n int32) bool { return n == 0 })
var op_jpg = mkctrl_op(func(n int32) bool { return n > 0 })
var op_jpl = mkctrl_op(func(n int32) bool { return n < 0 })

func op_int(i Instruction, m *Machine) (Interrupt, bool) {
	v := m.Text(m.PC() + 1)
	m.AddPC(1)
	return Interrupt(v), true
}

func op_jmp(i Instruction, m *Machine) (Interrupt, bool) {
	v := uint32(m.Text(m.PC() + 1))
	v = uint32(m.Text(m.PC()+2)) << 16
	m.SetPC(v)
	return INT_ZERO, false
}

func op_loa(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	v := uint32(m.Text(m.PC() + 1))
	v = uint32(m.Text(m.PC()+2)) << 16
	x.Set(Word(m.Text(v)))
	m.AddPC(2)
	return INT_ZERO, false
}

func op_set(i Instruction, m *Machine) (Interrupt, bool) {
	x := m.R(i.X())
	v := uint32(m.Text(m.PC() + 1))
	v = uint32(m.Text(m.PC()+2)) << 16
	x.Set(Word(v))
	m.AddPC(2)
	return INT_ZERO, false
}

var OpFuncs []OpFunc = []OpFunc{
	OP_MOV: op_mov,
	OP_MVZ: op_mvz,
	OP_PSH: op_psh,
	OP_POP: op_pop,
	OP_PAL: op_pal,
	OP_PEK: op_pek,
	OP_ADD: op_add,
	OP_ADX: op_adx,
	OP_SUB: op_sub,
	OP_SBX: op_sbx,
	OP_MUL: op_mul,
	OP_DIV: op_div,
	OP_MOD: op_mod,
	OP_AND: op_and,
	OP_BOR: op_bor,
	OP_XOR: op_xor,
	OP_SHL: op_shl,
	OP_SHR: op_shr,
	OP_XAD: op_xad,
	OP_XSB: op_xsb,
	OP_JPN: op_jpn,
	OP_JPZ: op_jpz,
	OP_JPG: op_jpg,
	OP_JPL: op_jpl,
	OP_INT: op_int,
	OP_JMP: op_jmp,
	OP_LOA: op_loa,
	OP_SET: op_set,
}
