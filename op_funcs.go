package rhmrm

type OpFunc func(m *Machine, args ...Word)

// get 1 arg
func get1(args []Word) Word {
	return args[0]
}

// get 2 args
func get2(args []Word) (Word, Word) {
	return args[0], args[1]
}

// get 2 registers
func get2r(args []Word, m *Machine) (*Word, *Word) {
	return m.R(args[0]), m.R(args[1])
}

// get register and the next word
func get2rn(args []Word, m *Machine) (*Word, Word) {
	return m.R(args[0]), *m.Mem(*m.PC())
}

/*
// radds does add signed 10-bit value to a register (duude, that's rad)
func raddc(r *Word, v Word) {
	*r = Word(int16(*r) + int16(sextend10(v)))
}

// jump (relative) by sign-extended 10-bit v
func jump(v Word) {
	*m.PC() += sextend10(v)
}
*/

// Ordinary operators.
var op_funcs = []OpFunc{
	OP_IMP: func(m *Machine, args ...Word) {
		a, b := get2(args)
		imp_funcs[a](m, b)
		*m.PC()++
	},
	OP_MOV: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a = *b
	},
	OP_MTC: func(m *Machine, args ...Word) {
		k, b := get2(args)
		md := k >> 3
		kr := k & 7

		switch md {
		case AM_SET: *m.C(kr) = *m.R(b)
		case AM_AND: *m.C(kr) &= *m.R(b)
		case AM_IOR: *m.C(kr) |= *m.R(b)
		case AM_XOR: *m.C(kr) ^= *m.R(b)
		}
	},
	OP_MFC: func(m *Machine, args ...Word) {
		a, k := get2(args)
		md := k >> 3
		kr := k & 7

		switch md {
		case AM_SET: *m.R(a) = *m.C(kr)
		case AM_AND: *m.R(a) &= *m.C(kr)
		case AM_IOR: *m.R(a) |= *m.C(kr)
		case AM_XOR: *m.R(a) ^= *m.C(kr)
		}
	},
	OP_STR: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*m.Mem(*a) = *b
	},
	OP_PSH: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a--
		*m.Mem(*a) = *b
	},
	OP_LOA: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a = *m.Mem(*b)
	},
	OP_POP: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a = *m.Mem(*b)
		*b++
	},
	OP_MOM: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*m.Mem(*a) = *m.Mem(*b)
		*a++
		*b++
	},

	OP_SRL: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a = *m.PC()
		*m.PC() = *b
	},
	
	OP_ADD: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) + uint32(*b)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_ADX: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) + uint32(*b) + uint32(*ex)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_SUB: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) - uint32(*b)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_SBX: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) - uint32(*b) + uint32(*ex)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_MUL: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) * uint32(*b)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_MLI: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := int32(*a) * int32(*b)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_DIV: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) << 16 / uint32(*b)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	OP_DVI: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := int32(*a) << 16 / int32(*b)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	OP_MOD: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a %= *b
	},
	OP_MDI: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a = Word(int16(*a) % int16(*b))
	},
	OP_INC: func(m *Machine, args ...Word) {
		a, b := m.R(args[0]), args[1]
		*a += b
	},

	OP_AND: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a &= *b
	},
	OP_IOR: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a |= *b
	},
	OP_XOR: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		*a ^= *b
	},
	OP_SHL: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) << uint32(*b)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	OP_ASR: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := int32(*a) << 16 >> uint32(*b)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	OP_SHR: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) << 16 >> uint32(*b)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	OP_ROL: func(m *Machine, args ...Word) {
		a, b := m.R(args[0]), args[1]
		r := uint32(*a) << uint32(b)
		*a = Word(r) | Word(r >> 16)
	},
	OP_ROR: func(m *Machine, args ...Word) {
		a, b := m.R(args[0]), args[1]
		r := uint32(*a) << 16 >> uint32(b)
		*a = Word(r) | Word(r >> 16)
	},

	OP_TST: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		*ex = *a & *b
	},
	OP_TEQ: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		*ex = *a ^ *b
	},
	OP_CMP: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		*ex = *a - *b
	},
	OP_CMN: func(m *Machine, args ...Word) {
		a, b := get2r(args, m)
		ex := m.C(C_EX)
		*ex = *a + *b
	},

	OP_JMP: func(m *Machine, args ...Word) {
		c := args[0]
		*m.PC() += sextend10(c) - 1
	},
	OP_JLT: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex < 0 {
			*m.PC() += sextend10(c) - 1
		}
	},
	OP_JLE: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex <= 0 {
			*m.PC() += sextend10(c) - 1
		}
	},
	OP_JGT: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex > 0 {
			*m.PC() += sextend10(c) - 1
		}
	},
	OP_JGE: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex >= 0 {
			*m.PC() += sextend10(c) - 1
		}
	},
	OP_JEQ: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex == 0 {
			*m.PC() += sextend10(c) - 1
		}
	},
	OP_JNE: func(m *Machine, args ...Word) {
		c := args[0]
		ex := m.C(C_EX)
		if *ex != 0 {
			*m.PC() += sextend10(c) - 1
		}
	},

	OP_SWI: func(m *Machine, args ...Word) {	
		*m.C(C_IR) = *m.PC()
		*m.PC() = *m.C(C_IA)
		*m.C(C_IM) = args[0]
		(*FlagsRegister)(m.C(C_FL)).SetI(true)
		(*FlagsRegister)(m.C(C_FL)).SetS(true)
	},
	OP_HWI: func(m *Machine, args ...Word) {
		m.interrupt.trigger = true
		m.interrupt.message = args[0]
		*m.C(C_IM) = args[0]
	},
	OP_IRE: func(m *Machine, args ...Word) {
		*m.PC() = *m.C(C_IR)
		(*FlagsRegister)(m.C(C_FL)).SetI(false)
		if args[0] != 0 {
			(*FlagsRegister)(m.C(C_FL)).SetS(false)
		}
	},
}

// Immediate operand operators
var imp_funcs = []OpFunc{
	IMP_BRK: func(m *Machine, args ...Word) {
		*m.C(C_IR) = *m.PC()
		*m.PC() = *m.C(C_IA)
		*m.C(C_IM) = 0xffff
		(*FlagsRegister)(m.C(C_FL)).SetI(true)
		(*FlagsRegister)(m.C(C_FL)).SetS(true)
	},
	IMP_MOV: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a = n
	},
	IMP_MTC: func(m *Machine, args ...Word) {
		k, n := get2(args)
		md := k >> 3
		kr := k & 7

		switch md {
		case AM_SET: *m.C(kr) = n
		case AM_AND: *m.C(kr) &= n
		case AM_IOR: *m.C(kr) |= n
		case AM_XOR: *m.C(kr) ^= n
		}
	},

	IMP_STR: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*m.Mem(*a) = n
	},
	IMP_PSH: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a--
		*m.Mem(*a) = n
	},

	IMP_SRL: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a = *m.PC()
		*m.PC() = n - 1 // compensate OP_IMP incrementing PC by one
	},

	IMP_ADD: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) + uint32(n)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_ADX: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) + uint32(n) + uint32(*ex)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_SUB: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) - uint32(n)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_SBX: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) - uint32(n) + uint32(*ex)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_MUL: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) * uint32(n)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_MLI: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := int32(*a) * int32(n)
		*a, *ex = Word(r), Word(r >> 16)
	},
	IMP_DIV: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := uint32(*a) << 16 / uint32(n)
		*a, *ex = Word(r >> 16), Word(r)
	},
	IMP_DVI: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		r := int32(*a) << 16 / int32(n)
		*a, *ex = Word(r >> 16), Word(r)
	},
	IMP_MOD: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a %= n
	},
	IMP_MDI: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a = Word(int16(*a) % int16(n))
	},
	IMP_INC: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a += n
	},

	IMP_AND: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a &= n
	},
	IMP_IOR: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a |= n
	},
	IMP_XOR: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a ^= n
	},
	IMP_BIC: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		*a &^= n
	},
	IMP_SHL: func(m *Machine, args ...Word) {
		a, n, ex := m.R(args[0]), args[1], m.C(C_EX)
		r := uint32(*a) << uint32(n)
		*ex = Word(r >> 16)
		*a = Word(r)
	},
	IMP_ASR: func(m *Machine, args ...Word) {
		a, n, ex := m.R(args[0]), args[1], m.C(C_EX)
		r := int32(*a) << 16 >> uint32(n)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	IMP_SHR: func(m *Machine, args ...Word) {
		a, n, ex := m.R(args[0]), args[1], m.C(C_EX)
		r := uint32(*a) << 16 >> uint32(n)
		*ex = Word(r)
		*a = Word(r >> 16)
	},
	IMP_ROL: func(m *Machine, args ...Word) {
		a, n := m.R(args[0]), args[1]
		r := uint32(*a) << uint32(n)
		*a = Word(r) | Word(r >> 16)
	},
	IMP_ROR: func(m *Machine, args ...Word) {
		a, n := m.R(args[0]), args[1]
		r := uint32(*a) << 16 >> uint32(n)
		*a = Word(r) | Word(r >> 16)
	},

	IMP_TST: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		*ex = *a & n
	},
	IMP_TEQ: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		*ex = *a ^ n
	},
	IMP_CMP: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		*ex = *a - n
	},
	IMP_CMN: func(m *Machine, args ...Word) {
		a, n := get2rn(args, m)
		ex := m.C(C_EX)
		*ex = *a + n
	},
}
