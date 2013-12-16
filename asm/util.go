package asm

import "strconv"

import "github.com/niksaak/rhmrm/machine"

var control_regs = map[string]int{
	"pc": machine.PC,
	"ex": machine.EX,
	"ia": machine.IA,
	"im": machine.IM,
	"ir": machine.IR,
	"fl": machine.FL,
}
var general_regs = map[string]int{
	"zr": machine.ZR,
	"ra": machine.RA,
	"fp": machine.FP,
	"sp": machine.SP,
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
	n, err := strconv.ParseInt(r[1:], 10, 8)
	if err != nil {
		return 0, 0, false
	}
	switch r[0] {
	case 'c':
		return regnchk(7, cr, n)
	case 'r':
		return regnchk(31, gr, n)
	case 's':
		return regnchk(machine.S+7, gr, machine.S+n)
	case 't':
		return regnchk(machine.T+7, gr, machine.T+n)
	case 'v':
		return regnchk(machine.V+3, gr, machine.V+n)
	case 'a':
		return regnchk(machine.A+7, gr, machine.A+n)
	}
	return 0, 0, false
}

// get base from string, return base and string without base part.
func getBase(s string) (int, string) {
	// strings of length 1 do not have any affixes
	if len(s) == 1 {
		return 10, s
	}
	// check for suffix
	switch suf, end := s[len(s)-1:], len(s)-1; suf {
	case "h", "x":
		return 16, s[:end]
	case "o":
		return 8, s[:end]
	case "b":
		return 2, s[:end]
	}
	// check for prefix
	switch pre := s[:2]; pre {
	case "0x", "0h":
		return 16, s[2:]
	case "0o":
		return 8, s[2:]
	case "0b":
		return 2, s[2:]
	}
	// otherwise base is 10
	return 10, s
}

// convert string to an integer.
func atoi(s string) (int, error) {
	b, s := getBase(s)
	n, err := strconv.ParseInt(s, b, 32)
	return int(n), err
}
