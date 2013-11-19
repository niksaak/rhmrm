package asm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

// Lexeme types.
const (
	EOF = -(iota + 1)
	Symbol
	Integer
	Rune
	String
	Comment
)

var lexeme_string = map[rune]string{
	EOF: "EOF",
	Symbol: "Symbol",
	Integer: "Integer",
	Rune: "Rune",
	String: "String",
	Comment: "Comment",
}

// LexemeString takes lexeme or rune and returns a printable string.
func LexemeString(lm rune) string {
	if s, found := lexeme_string[lm]; found {
		return s
	}
	return fmt.Sprintf("%q", string(lm))
}

// Position in source, valid if Line > 0.
type Position struct {
	File   string // file name
	Offset int    // byte offset
	Line   int    // line number, starting at 1
	Column int    // column number, starting at 1
}

// ValidP returns true if Position is valid.
func (p *Position) ValidP() bool {
	return p.Line > 0
}

// String representation of Position
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

const buf_len = 1024

// Lexer implements reading of Unicode characters and tokens from io.Reader.
type Lexer struct {
	// Input
	rd io.Reader // file reader

	// Source buffer
	rd_buf [buf_len + 1]byte // character buffer
	rd_pos int               // buffer position in file
	rd_end int               // buffer end position in file

	// Buffer position
	rd_buf_offset int // offset of rd_buf[0] in source
	line          int // line count
	column        int // character count
	last_line_len int // length of last line in characters
	last_char_len int // length of last character in bytes

	// Lexeme text buffer
	lm_buf bytes.Buffer // lexeme text head head that is not in rd_buf
	lm_pos int          // lexeme text tail position
	lm_end int          // lexeme text tail end

	// One character look-ahead
	ch rune // character before current rd_pos

	// Error is called for each error encountered
	ErrorCount int // incremented by one for each error occured

	// Start position of most recently scanned token.
	// Init and Next invalidate this.
	Position
}

// Init initializes a Lexer
func (l *Lexer) Init(src io.Reader) *Lexer {
	l.rd = src

	// init read buffer
	// (first call to next will fill it calling rd.Read)
	l.rd_buf[0] = utf8.RuneSelf // sentinel
	l.rd_pos = 0
	l.rd_end = 0

	// init source position
	l.rd_buf_offset = 0
	l.line = 1
	l.column = 0
	l.last_line_len = 0
	l.last_char_len = 0

	// init lexeme text buffer
	// (required for the first call to next())
	l.lm_pos = -1

	// init one character look-ahead
	l.ch = -1

	// init public fields
	l.ErrorCount = 0

	l.Line = 0

	return l
}

// next reads and returns the next unicode character
// TODO: I am sure it can be optimized (like, even more).
func (l *Lexer) next() rune {
	ch, width := rune(l.rd_buf[l.rd_pos]), 1

	if ch >= utf8.RuneSelf {
		// not ASCII or not enough bytes
		for l.rd_pos+utf8.UTFMax > l.rd_end &&
			!utf8.FullRune(l.rd_buf[l.rd_pos:l.rd_end]) {
			// not enough bytes: go to the store and read some more
			// but first save away lexeme text, if any
			if l.lm_pos >= 0 {
				l.lm_buf.Write(l.rd_buf[l.lm_pos:l.rd_pos])
				l.lm_pos = 0
				// lm_end is set by Scan()
			}
			// move unread bytes to the beginning of the buffer
			copy(l.rd_buf[0:], l.rd_buf[l.rd_pos:l.rd_end])
			l.rd_buf_offset += l.rd_pos
			// read more bytes
			i := l.rd_end - l.rd_pos
			n, err := l.rd.Read(l.rd_buf[i:buf_len])
			l.rd_pos = 0
			l.rd_end = i + n
			l.rd_buf[l.rd_end] = utf8.RuneSelf // sentinel
			if err != nil {
				if l.rd_end == 0 {
					if l.last_char_len > 0 {
						// previous char was not EOF
						l.column++
					}
					return EOF
				}
				if err != io.EOF {
					l.error(err.Error())
				}
				// if err == EOF, we won't be getting more
				// bytes; break to avoid infinite loop. If
				// err is something else, we don't know if
				// we can get more bytes, thus also break.
				break
			}
		}
		// at least one byte
		ch = rune(l.rd_buf[l.rd_pos])
		if ch >= utf8.RuneSelf {
			// not ASCII
			ch, width = utf8.DecodeRune(l.rd_buf[l.rd_pos:l.rd_end])
			if ch == utf8.RuneError && width == 1 {
				// advance for correct error position
				l.rd_pos += width
				l.last_char_len = width
				l.column++
				l.error("illegal UTF-8 encoding")
				return ch
			}
		}
	}
	// advance
	l.rd_pos += width
	l.last_char_len = width
	l.column++

	// special situations
	switch ch {
	case 0:
		// don't accept NUL
		l.error("illegal character NUL")
	case '\n':
		l.line++
		l.last_line_len = l.column
		l.column = 0
	}

	return ch
}

// Next reads and returns the next Unicode character.
// It returns EOF at the end of the source. Next does
// not update the Scanner position, for that use Pos.
func (l *Lexer) Next() rune {
	l.lm_pos = -1 // don't collect token text
	l.Line = 0 // invalidate token position
	ch := l.Peek()
	l.ch = l.next()
	return ch
}

// Peek returns the next Unicode character without advancing
// the scanner. It returns EOF if the Scanner position is at
// the last character in the source.
func (l *Lexer) Peek() rune {
	if l.ch < 0 {
		// first time reading
		l.ch = l.next()
		if l.ch == '\uFEFF' {
			l.ch = l.next()
		}
	}
	return l.ch
}

// Pos returns position of character immediately after
// the character or lexeme returned by the last call to Next or Scan.
func (l *Lexer) Pos() (p Position) {
	p.File = l.File
	p.Offset = l.rd_buf_offset + l.rd_pos - l.last_char_len
	switch {
	case l.column > 0:
		// last character was not a '\n'
		p.Line = l.line
		p.Column = l.column
	case l.last_line_len > 0:
		// last character was a '\n'
		p.Line = l.line - 1
		p.Column = l.last_line_len
	default:
		// at the beginning of the source
		p.Line = 1
		p.Column = 1
	}
	return
}

// Scan reads the next lexeme or Unicode character from source and returns it.
// It returns EOF at the end of the source.
func (l *Lexer) Scan() (rune, string) {
	ch := l.Peek()

	// reset lexeme text pos
	l.lm_pos = -1
	l.Line = 0

	// skip whitespace
	for whitespacep(ch) {
		ch = l.next()
	}

	// start collecting lexeme text
	l.lm_buf.Reset()
	l.lm_pos = l.rd_pos - l.last_char_len

	// set lexeme position
	l.Position = l.Pos()

	// determine lexeme value
	lm := ch
	switch {
	case unicode.IsLetter(ch) || ch == '_':
		lm = Symbol
		ch = l.scan_symbol()
	case decimalp(ch):
		lm = Integer
		ch = l.scan_integer(ch)
	default:
		switch ch {
		case '"':
			lm = String
			l.scan_string(ch)
			ch = l.next()
		case '\'':
			lm = Rune
			l.scan_rune()
			ch = l.next()
		case ';':
			lm = Comment
			ch = l.scan_comment()
		default:
			ch = l.next()
		}
	}

	// end of lexeme text
	l.lm_end = l.rd_pos - l.last_char_len

	l.ch = ch
	return lm, l.lexeme_text()
}

// scan_symbol tokenizes a symbol.
func (l *Lexer) scan_symbol() rune {
	ch := l.next()
	for ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) {
		ch = l.next()
	}
	return ch
}

// scan_integer tokenizes an integer.
// Note: actually tokenizes everything until encountering first whitespace.
//       Because pain. Validate at AST-building stage.
func (l *Lexer) scan_integer(ch rune) rune {
	ch = l.next()
	for !whitespacep(ch) && ch >= 0 {
		ch = l.next()
	}
	return ch
}

// scan_string tokenizes a string.
func (l *Lexer) scan_string(quot rune) (n int) {
	ch := l.next()
	for ch != quot {
		if ch < 0 {
			l.error("literal not terminated")
		}
		if ch == '\\' {
			ch = l.scan_escape(quot)
		} else {
			ch = l.next()
		}
		n++
	}
	return
}

// scan_escape skips character escape.
func (l *Lexer) scan_escape(quot rune) rune {
	ch := l.next()
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quot:
		ch = l.next()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		ch = l.scan_digits(ch, 8, 3)
	case 'x':
		ch = l.scan_digits(l.next(), 16, 2)
	case 'u':
		ch = l.scan_digits(l.next(), 16, 4)
	case 'U':
		ch = l.scan_digits(l.next(), 16, 8)
	default:
		l.error("illegal char escape")
	}
	return ch
}

// scan_digits skips up to n digits in base
func (l *Lexer) scan_digits(ch rune, base, n int) rune {
	for n > 0 && digit_value(ch) < base {
		ch = l.next()
		n--
	}
	if n > 0 {
		l.error("illegal rune escape")
	}
	return ch
}

// scan_rune tokenizes one rune literal.
func (l *Lexer) scan_rune() {
	if l.scan_string('\'') != 1 {
		l.error("illegal rune literal")
	}
}

// scan_comment tokenizes a comment.
func (l *Lexer) scan_comment() rune {
	ch := l.next()
	for ch != '\n' && ch >= 0 {
		ch = l.next()
	}
	return ch
}

// lexeme_text returns text for last lexeme scanned.
func (l *Lexer) lexeme_text() string {
	if l.lm_pos < 0 {
		// no token text
		return ""
	}
	if l.lm_end < 0 {
		// if EOF is reached, l.lm_end is set to -1, l.rd_pos == 0
		l.lm_end = l.lm_pos
	}
	if l.lm_buf.Len() == 0 {
		// the entire lexeme text is still in rd_buf
		return string(l.rd_buf[l.lm_pos:l.lm_end])
	}
	// part of the lexeme text is saved in lm_buf
	l.lm_buf.Write(l.rd_buf[l.lm_pos:l.lm_end])
	l.lm_pos = l.lm_end
	return l.lm_buf.String()
}

// error increments error count and calls l.Error if present,
// then printing the error string to Stderr.
func (l *Lexer) error(msg string) {
	l.ErrorCount++
	pos := l.Position
	if !pos.ValidP() {
		pos = l.Pos()
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", pos, msg)
}

// decimalp returns true if ch is a decimal digit.
func decimalp(ch rune) bool {
	return unicode.IsDigit(ch)
}

// digit_value returns integer value for hexadecimal digit.
func digit_value(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

// whitespacep returns true if ch is whitespace.
func whitespacep(ch rune) bool {
	return unicode.IsSpace(ch)
}
