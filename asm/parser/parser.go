// Package parser implements generating parse trees from lexers.
package parser

import (
	"fmt"
	"github.com/niksaak/rhmrm/asm/compiler"
	"github.com/niksaak/rhmrm/asm/lexer"
	"github.com/niksaak/rhmrm/asm/util"
	"unicode"
)

type lexeme struct {
	pos lexer.Position // position info
	k   rune           // kind
	lit string         // literal string which lexeme represents
}

// Parser implements building abstract syntax tree from lexeme stream.
type Parser struct {
	lexer      *lexer.Lexer
	lexeme     // current lexeme
	ErrorCount int
}

func (p *Parser) Init(lx *lexer.Lexer) *Parser {
	// initialize values
	p.lexer = lx
	p.ErrorCount = 0

	p.next() // fetch the first token
	return p
}

func (p *Parser) next() {
	p.pos, p.k, p.lit = p.lexer.Scan()
}

// ParseProgram transforms source into a tree.
func (p *Parser) ParseProgram() compiler.Node {
	prog := new(compiler.ProgramNode)
	switch p.k {
	case lexer.ILL:
		if p.lexer == nil {
			return p.error("parser is not initialized")
		} else {
			return p.errorf("illegal lexeme: %s", p.lit)
		}
	}
	for p.k != lexer.EOF {
		nodes := p.parseClause()
		prog.Clauses = ndappend(prog.Clauses, nodes...)
	}
	return prog
}

// clause = [ label ] [ instruction | directive ] [ comment ] "\n" .
func (p *Parser) parseClause() (nodes []compiler.Node) {
	// parse label
	nodes = append(nodes, p.parseLabel())

	// parse directive or instruction
	if n := p.parseDirective(); n != nil {
		nodes = append(nodes, n)
	} else if n := p.parseInstruction(); n != nil {
		nodes = append(nodes, n)
	}

	// parse comment
	nodes = append(nodes, p.parseComment())

	// check for errors
	if p.k != '\n' {
		if p.k == lexer.EOF {
			// illegal EOF is already reported at this point
			return
		}
		nodes = append(nodes,
			p.errorf("unrecognized lexeme: "+p.lit))
		p.next()
		return p.parseClause()
	}
	p.next()
	return
}

// label = ":" symbol .
func (p *Parser) parseLabel() compiler.Node {
	if p.k != ':' { // labels start with colon
		return nil
	}
	pos := p.pos
	p.next()
	if err := p.lmexpect(lexer.SYMBOL); err != nil {
		return err
	}
	name := p.lit
	p.next()
	return &compiler.LabelNode{pos, name}
}

// directive = "." symbol [ operands ] .
func (p *Parser) parseDirective() compiler.Node {
	if p.k != '.' { // directives start with a dot
		return nil
	}
	pos := p.pos
	p.next()

	// operator
	if err := p.lmexpect(lexer.SYMBOL); err != nil {
		return err
	}
	sym := p.lit
	p.next()

	// operands
	operands := p.parseOperands()

	return &compiler.DirectiveNode{pos, sym, operands}
}

// instruction = symbol [ operands ] .
func (p *Parser) parseInstruction() compiler.Node {
	if p.k != lexer.SYMBOL { // instructions start with a symbol
		return nil
	}
	pos := p.pos

	// operator
	sym := p.lit
	p.next()

	// operands
	operands := p.parseOperands()

	return &compiler.InstructionNode{pos, sym, operands}
}

// comment = <comment-line-token> .
func (p *Parser) parseComment() (c compiler.Node) {
	if p.k != lexer.COMMENT { // comments are termins
		return nil
	}
	i := 0
	level := 0
	// get comment level from semicolon count
	for i < len(p.lit) && p.lit[i] == ';' {
		level++
		i++
	}
	// skip leftover whitespace
	// TODO: make checking for whitespace more efficient.
	for i < len(p.lit) && unicode.IsSpace(rune(p.lit[i])) {
		i++
	}
	c = &compiler.CommentNode{p.pos, level, p.lit[i:]}
	p.next()
	return
}

// operands = [ operand { [ "," ] operand } ] .
func (p *Parser) parseOperands() (operands []compiler.Node) {
	o := p.parseOperand()
	for o != nil {
		if p.k == ',' { // skip commas
			p.next()
		}
		operands = append(operands, o)
		o = p.parseOperand()
	}
	return
}

// operand = register | symbol | integer | string | block .
func (p *Parser) parseOperand() (o compiler.Node) {
	// pro'lly there's a better way, but this looks kinda cool too
	for _, fn := range []func() compiler.Node{
		p.parseRegister,
		p.parseSymbol,
		p.parseInteger,
		p.parseString,
		p.parseBlock,
	} {
		o = fn()
		if o != nil {
			// p.next() called by fn()
			return
		}
	}
	return nil
}

// register = <general-register> | [ <access-mode> ] <control-register> .
func (p *Parser) parseRegister() (er compiler.Node) {
	r := &compiler.RegisterNode{Position: p.pos}
	switch p.k {
	case '&', '|', '^':
		// access modes for control register.
		r.Kind = util.ControlRegisterKind & util.ControlModes[p.k]
		p.next()
		if err := p.lmexpect(lexer.REGISTER); err != nil {
			return nil
		}
		k, n, ok := util.Reginfo(p.lit)
		switch {
		case !ok:
			er = p.errorf("bad register: %s (%d,%d)", p.lit, k, n)
			return
		case k != util.ControlRegisterKind:
			er = p.errorf("register is not control: %s", p.lit)
			return
		}
		r.Index = n
	case '=':
		p.next()
		fallthrough
	default:
		if p.k != lexer.REGISTER {
			return nil
		}
		// TODO: get rid of repeating somehow.
		k, n, ok := util.Reginfo(p.lit)
		if !ok {
			er = p.errorf("bad register: %s (%d,%d)", p.lit, k, n)
			return
		}
		r.Kind = k
		r.Index = n
	}
	p.next()
	return r
}

// symbol = <symbol> .
func (p *Parser) parseSymbol() (n compiler.Node) {
	if p.k != lexer.SYMBOL {
		return nil
	}
	n = &compiler.SymbolNode{p.pos, p.lit}
	p.next()
	return
}

// integer = <integer, prefixed or suffixed>
func (p *Parser) parseInteger() compiler.Node {
	if p.k != lexer.INTEGER {
		return nil
	}
	n, err := util.Atoi(p.lit)
	if err != nil {
		return p.errorf("bad integer: %s (%s)", p.lit, err)
	}
	p.next()
	return &compiler.IntegerNode{p.pos, n}
}

// string = '"' <anything> '"'
func (p *Parser) parseString() (n compiler.Node) {
	if p.k != lexer.STRING {
		return nil
	}
	n = &compiler.StringNode{p.pos, p.lit}
	p.next()
	return
}

// block = "{" { clause } "}"
func (p *Parser) parseBlock() (n compiler.Node) {
	if p.k != '{' { // blocks start with '{'
		return nil
	}
	p.next()
	b := new(compiler.BlockNode)
	if p.k == lexer.COMMENT { // accept comment after '{'
		b.Clauses = append(b.Clauses, p.parseComment())
		p.next()
	}
	/*
		// FIXME: why do we not get '\n' here?
		if err := p.lmexpect('\n'); err != nil {
			return err
		}
	*/
	for p.k != '}' {
		if p.k == lexer.EOF {
			return p.error("unexpected EOF")
		}
		b.Clauses = ndappend(b.Clauses, p.parseClause()...)
	}
	b.Position = p.pos
	p.next()
	return b
}

// error returns Errorcompiler.Node with current position and supplied message.
func (p *Parser) error(msg string) *compiler.ErrorNode {
	p.ErrorCount++
	return &compiler.ErrorNode{p.pos, msg}
}

// errorf is like error with format.
func (p *Parser) errorf(msg string, args ...interface{}) *compiler.ErrorNode {
	return p.error(fmt.Sprintf(msg, args...))
}

// lmexpect is a convenient helper for checking if current lexeme is desired.
func (p *Parser) lmexpect(lm rune) (err *compiler.ErrorNode) {
	if p.k != lm {
		return p.errorf("expected %s, got %q (%s)",
			lexer.LexemeString(lm), p.lit, lexer.LexemeString(p.k))
	}
	return nil
}

// append non-nil nodes to a slice and return it.
func ndappend(slice []compiler.Node, nodes ...compiler.Node) []compiler.Node {
	if nodes == nil {
		return slice
	}
	for _, n := range nodes {
		if n != nil {
			slice = append(slice, n)
		}
	}
	return slice
}
