package asm

import "strconv"

import . "github.com/niksaak/rhmrm"

var control_regs = map[string]int{
	"pc": C_PC,
	"ex": C_EX,
	"ia": C_IA,
	"im": C_IM,
	"ir": C_IR,
	"fl": C_FL,
}
var general_regs = map[string]int{
	"zr": R_ZR,
	"ra": R_RA,
	"fp": R_FP,
	"sp": R_SP,
}

// check register index.
func regnchk(max, kind int, num int64) (int, int, bool) {
	if int(num) > max || int(num) < 0 {
		return 0, 0, false
	}
	return kind, int(num), true
}

// check if string designates register and return register kind and index.
func reginfo(r string) (kind, num int, ok bool) {
	const cr = controlRegisterKind
	const gr = generalRegisterKind
	if n, found := control_regs[r]; found {
		return cr, n, true
	}
	if n, found := general_regs[r]; found {
		return gr, n, true
	}
	k := r[0]
	n, err := strconv.ParseInt(r, 10, 8)
	if err != nil {
		return 0, 0, false
	}
	switch k {
	case 'c':
		return regnchk(7, cr, n)
	case 'r':
		return regnchk(31, gr, n)
	case 's':
		return regnchk(R_S7, gr, R_S0+n)
	case 't':
		return regnchk(R_T7, gr, R_T0+n)
	case 'v':
		return regnchk(R_V3, gr, R_V0+n)
	case 'a':
		return regnchk(R_A7, gr, R_A0+n)
	}
	return 0, 0, false
}

// get base from string, return base and string without base part.
func getBase(s string) (int, string) {
	switch suf, end := s[len(s)-1:], len(s)-1; suf {
	case "h", "x":
		return 16, s[:end]
	case "o":
		return 8, s[:end]
	case "b":
		return 2, s[:end]
	}
	switch pre := s[:2]; pre {
	case "0x", "0h":
		return 16, s[2:]
	case "0o":
		return 8, s[2:]
	case "0b":
		return 2, s[2:]
	}
	return 10, s
}

// convert string to an integer.
func atoi(s string) (int, error) {
	b, s := getBase(s)
	n, err := strconv.ParseInt(s, b, 32)
	return int(n), err
}