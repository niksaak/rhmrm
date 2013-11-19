package asm

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type Streader struct {
	data []string
	step int
}

func (r *Streader) Read(p []byte) (n int, err error) {
	if r.step < len(r.data) {
		s := r.data[r.step]
		n = copy(p, s)
		r.step++
	} else {
		err = io.EOF
	}
	return
}

type lexeme struct {
	lm rune
	text string
}

var f100 = "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" +
	"fffffffffffffffffffffffffffffffffff"

var lexeme_list = []lexeme{
	{ Comment, "; line comments" },
	{ Comment, ";" },
	{ Comment, ";;" },
	{ Comment, ";comment" },
	{ Comment, "; comment;" },
	{ Comment, ";" + f100 },

	{ Comment, "; symbols" },
	{ Symbol, "a" },
	{ Symbol, "a0" },
	{ Symbol, "foobar" },
	{ Symbol, "abc123" },
	{ Symbol, "LGTM" },
	{ Symbol, "_" },
	{ Symbol, "_abc123" },
	{ Symbol, "_abc_123_" },
	{ Symbol, "_本" },
	{ Symbol, "本" },
	{ Symbol, "bar９８７６"},
	{ Symbol, f100 },

	{ Comment, "; integers" },
	{ Integer, "0" },
	{ Integer, "01" },
	{ Integer, "0x10c" },
	{ Integer, "0x" + f100 },
	{ Integer, "235h" },
	{ Integer, "0X" + f100 },

	{ Comment, "; runes" },
	{ Rune, `' '` },
	{ Rune, `'a'` },
	{ Rune, `'本'` },
	{ Rune, `'\a'` },
	{ Rune, `'\b'` },
	{ Rune, `'\f'` },
	{ Rune, `'\n'` },
	{ Rune, `'\r'` },
	{ Rune, `'\t'` },
	{ Rune, `'\v'` },
	{ Rune, `'\''` },
	{ Rune, `'\000'` },
	{ Rune, `'\777'` },
	{ Rune, `'\x00'` },
	{ Rune, `'\xff'` },
	{ Rune, `'\u0000'` },
	{ Rune, `'\ufA16'` },
	{ Rune, `'\U00000000'` },
	{ Rune, `'\U0000ffAB'` },

	{ Comment, "; strings" },
	{ String, `" "` },
	{ String, `"a"` },
	{ String, `"本"` },
	{ String, `"\a"` },
	{ String, `"\b"` },
	{ String, `"\f"` },
	{ String, `"\n"` },
	{ String, `"\r"` },
	{ String, `"\t"` },
	{ String, `"\v"` },
	{ String, `"\""` },
	{ String, `"\000"` },
	{ String, `"\777"` },
	{ String, `"\x00"` },
	{ String, `"\xff"` },
	{ String, `"\u0000"` },
	{ String, `"\ufA16"` },
	{ String, `"\U00000000"` },
	{ String, `"\U0000ffAB"` },
	{ String, `"` + f100 + `"` },

	{Comment, "; individual characters"},
	// NUL character is not allowed
	{ '\x01', "\x01" },
	{ ' ' - 1, string(' ' - 1) },
	{ '+', "+" },
	{ '/', "/" },
	{ '.', "." },
	{ '~', "~" },
	{ '(', "(" },
}

func mk_source(pattern string) *bytes.Buffer {
	var buf bytes.Buffer
	for _, lm := range lexeme_list {
		fmt.Fprintf(&buf, pattern, lm.text)
	}
	return &buf
}

func check_lexeme(t *testing.T, l *Lexer,
	line int, got, want rune, text string) {
	if got != want {
		t.Fatalf("tok=%s, want %s for %q",
			LexemeString(got), LexemeString(want), text)
	}
	if l.Line != line {
		t.Errorf("line=%d, want %d for %q", l.Line, line, text)
	}
	ltext := l.lexeme_text()
	if ltext != text {
		t.Errorf("text = %q, want %q", ltext, text)
	} // no need to check for idempotency because lexeme_text() is local
}

func count_new_lines(s string) (n int) {
	for _, ch := range s {
		if ch == '\n' {
			n++
		}
	}
	return
}

func TestScan(t *testing.T) {
	t.Parallel()
	l := new(Lexer).Init(mk_source(" \t%s\n"))
	lm, _ := l.Scan()
	line := 1
	for _, k := range lexeme_list {
		check_lexeme(t, l, line, lm, k.lm, k.text)
		lm, _ = l.Scan()
		line += count_new_lines(k.text) + 1
	}
	check_lexeme(t, l, line, lm, EOF, "")
}
