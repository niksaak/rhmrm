package lexer

import "testing"
import "fmt"

// ck_lexeme contains string and expected kind and literal.
type ck_lexeme struct {
	kind rune
	str  string
	lit  string
}

var checks = []ck_lexeme{
	{COMMENT, "; Comments", " Comments"},
	{COMMENT, "; Check \n newline", " Check "},

	{COMMENT, "; Symbols", ""},
	{SYMBOL, "foo", ""},
	{SYMBOL, "b4r", ""},
	{SYMBOL, "_ba2", ""},
	{SYMBOL, "Ghilber t", "Ghilber"},

	{COMMENT, "; Registers", ""},
	{REGISTER, "r0", ""},
	{REGISTER, "r31", ""},
	{REGISTER, "zr", ""},
	{REGISTER, "pc", ""},

	{COMMENT, "; Strings", ""},
	{STRING, `"Lorem ipsum dolor"`, "Lorem ipsum dolor"},
	{STRING, "\"Sit amet \n consectetur\"", "Sit amet \n consectetur"},

	{COMMENT, "; Integers", ""},
	{INTEGER, "0x10c", ""},
	{INTEGER, "800h", ""},
	{INTEGER, "9001this_is_invalid_for_parser_but_valid_for_lexer", ""},

	{COMMENT, "; Runes", ""},
	{RUNE, "'z'", "z"},
	{RUNE, "'ї'", "ї"},
	{RUNE, `'\x40'`, `\x40`},
	{RUNE, `'\u0407'`, `\u0407`},

	{'.', ".", "."},
	{':', ":", ""},
	{'{', "{", ""},
	{'}', "}", "}"},
	{'\n', "\n", ""},
}

func TestLexemeString(t *testing.T) {
	t.Parallel()
	ch := rune(EOF)
	if s := LexemeString(ch); s != "EOF" {
		t.Errorf(`LexemeString(EOF) is %q, want "EOF"`, s)
	}
	ch = '{'
	if s := LexemeString(ch); s != "{" {
		t.Errorf(`LexemeString({) is %q, want "{"`, s)
	}
}

func TestRegisterp(t *testing.T) {
	t.Parallel()
	r := "r13"
	if !registerp(r) {
		t.Errorf("registerp(%v) is false, want true", r)
	}
}

func test_expect_lexeme(fn func() (Position, rune, string),
	expect rune, t *testing.T) {
	_, k, lit := fn()
	if k != expect {
		t.Errorf("expected %s, got %q (%s)",
			LexemeString(expect), lit, LexemeString(k))
	}
}

func TestLexerEOF(t *testing.T) {
	t.Parallel()
	exp := []byte("next_token_must_be_EOF")
	lx := new(Lexer).Init(exp, "", mkteh(t))
	test_expect_lexeme(lx.Scan, SYMBOL, t)
	test_expect_lexeme(lx.Scan, EOF, t)
}

func TestLexerNewline(t *testing.T) {
	t.Parallel()
	exp := []byte("\nsymbol")
	lx := new(Lexer).Init(exp, "", mkteh(t))
	test_expect_lexeme(lx.Scan, '\n', t)
	test_expect_lexeme(lx.Scan, SYMBOL, t)
	test_expect_lexeme(lx.Scan, EOF, t)
}

func TestLexer(t *testing.T) {
	t.Parallel()
	lx := new(Lexer)
	for i := 0; i < len(checks); i++ {
		ck := checks[i]
		filename := fmt.Sprintf("test#%d", i)
		lx.Init([]byte(ck.str), filename, mkteh(t))
		_, lm, lit := lx.Scan()

		if lm != ck.kind {
			t.Errorf("bad kind for %q: got %s, want %s",
				ck.str,
				LexemeString(lm), LexemeString(ck.kind))
		}
		if ck.lit != "" && lit != ck.lit {
			t.Errorf("bad literal for %q: got %q, want %q",
				ck.str, lit, ck.lit)
		}
	}
}

// make testing error handler
func mkteh(t *testing.T) ErrorHandler {
	return func(pos Position, msg string) {
		t.Logf("lexer: %v: %s", &pos, msg)
	}
}
