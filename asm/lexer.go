package asm

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Lexeme kinds.
const (
	ILL rune = -iota // illegal character
	EOF              // end of file

	SYMBOL   // = <letter> { <letter> | <decimal> } .
	REGISTER // = <letter> ( ( "0" ... "31" ) | [ "1" ] <letter> )
	STRING   // = '"' <anything> '"' .
	INTEGER  // = <decimal> { <letter> | <decimal> } .
	RUNE     // = "'" <rune or character code> "'" .
	COMMENT  // = ";" <anything> .
)

var lxStrings = map[rune]string{
	ILL: "ILLEGAL",
	EOF: "EOF",
	SYMBOL: "symbol",
	REGISTER: "register",
	STRING: "string",
	INTEGER: "integer",
	RUNE: "rune",
	COMMENT: "comment",
}

func LexemeString(lm rune) (s string) {
	s, found := lxStrings[lm]
	if !found {
		s = string(lm)
	}
	return
}

// Position in source, valid if Line > 0.
type Position struct {
	File   string
	Offset int
	Line   int
	Column int
}

// ValidP is true when p is a valid Position.
func (p *Position) ValidP() bool {
	return p.Line > 0
}

// String returns a string in one of the several forms:
//   file:line:column    valid position with filename
//   line:column         valid position without filename
//   file                invalid position with filename
//   ???                 invalid position without filename
func (p *Position) String() (s string) {
	s = p.File
	if p.ValidP() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	if s == "" {
		s = "???"
	}
	return
}

var _ fmt.Stringer = new(Position)

type ErrorHandler func(p Position, msg string)

// Lexer for RHMRM assembly.
type Lexer struct {
	src []byte // source
	err ErrorHandler

	ch        rune // current character
	off       int  // character offset
	rd_off    int  // reading offset (position of the next character)

	ErrorCount int
	Position
}

const bom = 0xFEFF // byte order mark

// Init sets lexer to the initial state.
func (l *Lexer) Init(src []byte, filename string, err ErrorHandler) *Lexer {
	l.src = src
	l.err = err

	l.ch = ' '
	l.off = 0
	l.rd_off = 0

	l.ErrorCount = 0
	l.Position = Position{
		File:   filename,
		Offset: 0,
		Line:   1,
		Column: 1,
	}

	l.next()
	if l.ch == bom {
		l.next()
	}

	return l
}

// Read next unicode char into p.ch.
func (l *Lexer) next() {
	if l.rd_off < len(l.src) { // beginning or middle of the file
		l.off, l.Offset = l.rd_off, l.rd_off
		if l.ch == '\n' {
			l.Column = 0
			l.Line++
		}
		r, w := rune(l.src[l.rd_off]), 1
		switch {
		case r == 0:
			l.error(l.Position, "illegal character NUL")
		case r >= 0x80:
			// not ASCII
			r, w = utf8.DecodeRune(l.src[l.rd_off:])
			if r == utf8.RuneError && w == 1 {
				l.error(l.Position, "illegal UTF-8 encoding")
			} else if r == bom && l.off > 0 {
				l.error(l.Position, "illegal BOM")
			}
		}
		l.rd_off += w
		l.ch = r
		l.Column++
	} else { // end of file
		l.off = len(l.src)
		if l.ch == '\n' {
			l.Column = 0
			l.Line++
		}
		l.ch = EOF
	}
}

// Scan gets the next lexeme from source.
func (l *Lexer) Scan() (pos Position, lm rune, lit string) {
	l.skipWhitespace()

	// current token start
	pos = l.Position

	// determine token value
	switch {
	case letterp(l.ch):
		lit = l.scanSymbol()
		if registerp(lit) {
			lm = REGISTER
		} else {
			lm = SYMBOL
		}
	case decimalp(l.ch):
		// we resolve integers at parsing stage, so scanSymbol()
		// is applicable
		lit = l.scanSymbol()
		lm = INTEGER
	default:
		switch l.ch {
		case '"':
			lit = l.scanString()
			lm = STRING
		case '\'':
			lit = l.scanRune()
			lm = RUNE
		case ';':
			lit = l.scanComment()
			lm = COMMENT
		default:
			lit = string(l.ch)
			lm = l.ch
		}
	}
	return
}

func (l *Lexer) skipWhitespace() {
	for l.ch != '\n' && unicode.IsSpace(l.ch) {
		l.next()
	}
}

func (l *Lexer) scanSymbol() string {
	off := l.off // save position
	for letterp(l.ch) || decimalp(l.ch) {
		l.next()
	}
	return string(l.src[off:l.off])
}

func (l *Lexer) scanString() (s string) {
	pos := l.Position // save position
	l.next()          // skip the initial '"'
	off := l.off      // save offset

	for l.ch != '"' {
		ch := l.ch
		l.next()
		if ch == ILL || ch == EOF {
			l.error(pos, "string not terminated")
			break
		}
		if ch == '\\' {
			l.scanEscape('"')
		}
	}

	s = string(l.src[off : l.off])
	l.next()
	return
}

func (l *Lexer) scanRune() (s string) {
	pos := l.Position
	l.next() // skip the initial "'"
	off := l.off

	n := 0
	for l.ch != '\'' {
		ch := l.ch
		n++
		l.next()
		if ch == '\n' || ch == EOF || ch == ILL {
			l.error(pos, "character literal not terminated")
			n = 1
			break
		}
		if ch == '\\' {
			l.scanEscape('\'')
		}
	}

	if n != 1 {
		l.error(pos, "illegal character literal")
	}

	s = string(l.src[off : l.off])
	l.next()
	return
}

func (l *Lexer) scanComment() string {
	l.next() // skip the initial ';'
	off := l.off

	for l.ch != '\n' && l.ch >= 0 {
		l.next()
	}

	return string(l.src[off : l.off])
}

func (l *Lexer) scanEscape(quote rune) {
	pos := l.Position

	var i, base, max uint32
	switch l.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		// nothing to do
		l.next()
		return
	case '0', '1', '2', '3', '4', '5', '6', '7':
		i, base, max = 3, 8, 255
	case 'x':
		l.next()
		i, base, max = 2, 16, 255
	case 'u':
		l.next()
		i, base, max = 4, 16, unicode.MaxRune
	case 'U':
		l.next()
		i, base, max = 8, 16, unicode.MaxRune
	default:
		l.next() // allways make progress
		l.error(pos, "unrecognized escape sequence")
		return
	}

	var x uint32
	for ; i > 0 && l.ch != quote && l.ch >= 0; i-- {
		d := uint32(digitVal(l.ch))
		if d >= base {
			l.error(l.Position,
				"illegal character in escape sequence")
			break
		}
		x = x*base + d
		l.next()
	}
	// in case of error, consume remaining chars
	for ; i > 0 && l.ch != quote && l.ch >= 0; i-- {
		l.next()
	}
	if x > max || 0xD800 <= x && x < 0xE000 {
		l.error(pos,
			"escape sequence is invalid Unicode code point")
	}
}

func (l *Lexer) error(pos Position, msg string) {
	if l.err != nil {
		l.err(pos, msg)
	}
	l.ErrorCount++
}

func letterp(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func decimalp(ch rune) bool {
	return unicode.IsDigit(ch)
}

func registerp(s string) bool {
	if len(s) < 2 || 3 < len(s) {
		return false
	}
	switch s {
	case "zr", "ra", "fp", "sp", "pc", "ex", "ia", "im", "ir", "fl":
		return true
	default:
		if !letterp(rune(s[0])) {
			return false
		}
		for _, c := range s[1:] {
			if !decimalp(c) {
				return false
			}
		}
		return true
	}
	return false
}

func digitVal(ch rune) int {
	// FIXME: accept non-ASCII decimals.
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // invalid value
}
