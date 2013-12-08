package asm

import (
	"fmt"
)

type lexeme struct {
	pos Position // position info
	k   rune     // kind
	lit string   // literal string which lexeme represents
}

// Parser implements building abstract syntax tree from lexeme stream.
type Parser struct {
	err        ErrorHandler
	next       func() // function to get the next token, set by Init()
	lexeme            // current lexeme
	ErrorCount int
}

func (p *Parser) init(lx *Lexer, err ErrorHandler) *Parser {
	// initialize values
	p.err = err
	p.next = mkLexemeFetcher(lx, p)
	p.ErrorCount = 0

	p.next() // fetch the first token
	return p
}

func mkLexemeFetcher(lx *Lexer, p *Parser) func() {
	ch := make(chan lexeme, 128)

	// scanning sender
	go func() {
		var lm lexeme
		for lm.k != EOF {
			lm.pos, lm.k, lm.lit = lx.Scan()
			ch <- lm
		}
		close(ch)
	}()

	// fetching reseiver
	return func() {
		if lm, ok := <-ch; ok {
			p.lexeme = lm
		} else {
			p.lexeme = lexeme{k: EOF}
		}
	}
}

// ParseProgram transforms source into a tree.
func (p *Parser) ParseProgram() (prog ProgramNode) {
	switch p.k {
	case ILL:
		if p.next == nil {
			p.error("parser is not initialized")
		} else {
			p.errorf("illegal lexeme: %s", p.lit)
		}
	}
	for p.k != EOF {
		prog.clauses = ndappend(prog.clauses, p.parseClause()...)
	}
	return
}

// clause = [ label ] [ instruction | directive ] [ comment ] "\n" .
func (p *Parser) parseClause() (nodes []Node) {
	// parse label
	nodes = ndappend(nodes, p.parseLabel())

	// parse directive or instruction
	if n := p.parseDirective(); n != nil {
		nodes = append(nodes, n)
	} else if n := p.parseInstruction(); n != nil {
		nodes = append(nodes, n)
	}

	// parse comment
	nodes = ndappend(nodes, p.parseComment())

	// check for errors
	if p.k != '\n' {
		if p.k == EOF {
			// illegal EOF is already reported at this point
			return
		}
		p.errorf("unrecognized lexeme: " + p.lit)
		p.next()
		return p.parseClause()
	}
	p.next()
	return
}

// label = ":" symbol .
func (p *Parser) parseLabel() Node {
	if p.k != ':' { // labels start with colon
		return nil
	}
	pos := p.pos
	p.next()
	if !p.lmexpect(SYMBOL) {
		return nil
	}
	name := p.lit
	p.next()
	return &LabelNode{pos, name}
}

// directive = "." instruction .
func (p *Parser) parseDirective() Node {
	if p.k != '.' { // directives start with a dot
		return nil
	}
	pos := p.pos
	p.next()
	if !p.lmexpect(SYMBOL) {
		return nil
	}
	var operands []Node
	o := p.parseOperand()
	for o != nil {
		operands = append(operands, o)
		p.next()
	}
	return &DirectiveNode{pos, *p.parseSymbol().(*SymbolNode), operands}
}

// instruction = symbol [ operand { "," operand } ] .
func (p *Parser) parseInstruction() Node {
	if p.k != SYMBOL { // instructions start with a symbol
		return nil
	}
	pos := p.pos
	sym := *p.parseSymbol().(*SymbolNode)
	p.next()

	var operands []Node
	o := p.parseOperand()
	for o != nil {
		operands = append(operands, o)
		p.next()
	}

	return &InstructionNode{DirectiveNode{pos, sym, operands}}
}

// comment = <comment-line-token> .
func (p *Parser) parseComment() (c Node) {
	if p.k != COMMENT { // comments are atomic lexemes
		return nil
	}
	i := 0
	for i < len(p.lit) && p.lit[i] == ';' {
		i++
	}
	c = &CommentNode{p.pos, i, p.lit[i:]}
	p.next()
	return
}

// operand = register | symbol | integer | string | block .
func (p *Parser) parseOperand() (o Node) {
	// pro'lly there's a better way, but this looks kinda cool too
	for _, fn := range []func() Node{
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
func (p *Parser) parseRegister() Node {
	r := &RegisterNode{Position: p.pos}
	switch p.k {
	case '&', '|', '^':
		// access modes for control register.
		r.kind = controlRegisterKind & controlModes[p.k]
		p.next()
		if !p.lmexpect(SYMBOL) {
			return nil
		}
		k, n, ok := reginfo(p.lit)
		switch {
		case !ok:
			p.errorf("bad register: %s", p.lit)
			return nil
		case k != controlRegisterKind:
			p.errorf("register is not control: %s", p.lit)
			return nil
		}
		r.index = n
	case '=':
		p.next()
		fallthrough
	default:
		if !p.lmexpect(SYMBOL) {
			return nil
		}
		// TODO: get rid of repeating somehow.
		k, n, ok := reginfo(p.lit)
		if !ok {
			p.errorf("bad register: %s", p.lit)
			return nil
		}
		r.kind = k
		r.index = n
	}
	p.next()
	return r
}

// symbol = <symbol> .
func (p *Parser) parseSymbol() Node {
	if p.k != SYMBOL {
		return nil
	}
	p.next()
	return &SymbolNode{p.pos, p.lit}
}

// integer = <integer, prefixed or suffixed>
func (p *Parser) parseInteger() Node {
	if p.k != INTEGER {
		return nil
	}
	n, err := atoi(p.lit)
	if err != nil {
		p.errorf("bad integer: %s (%s)", p.lit, err)
		p.next()
	}
	p.next()
	return &IntegerNode{p.pos, n}
}

// string = '"' <anything> '"'
func (p *Parser) parseString() Node {
	if p.k != STRING {
		return nil
	}
	pos, str := p.pos, p.lit
	p.next()
	return &StringNode{pos, str}
}

// block = "{" { clause } "}"
func (p *Parser) parseBlock() Node {
	if p.k != '{' { // blocks start with '{'
		return nil
	}
	p.next()
	b := new(BlockNode)
	for p.k != '}' {
		if p.k == EOF {
			p.error("unexpected EOF")
			return nil
		}
		b.clauses = ndappend(b.clauses, p.parseClause()...)
	}
	b.Position = p.pos
	p.next()
	return b
}

// append non-nil nodes to a slice and return it.
func ndappend(slice []Node, nodes ...Node) []Node {
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

// error calls parser error handler with supplied message.
func (p *Parser) error(msg string) {
	if p.err != nil {
		p.err(p.pos, msg)
	}
	p.ErrorCount++
}

// errorf is like error with format.
func (p *Parser) errorf(format string, args ...interface{}) {
	p.error(fmt.Sprintf(format, args...))
}

// lmexpect is a convenient helper for checking if current lexeme is desired.
func (p *Parser) lmexpect(lm rune) bool {
	if p.k != lm {
		p.errorf("expected %s, got %s (%s)",
			LexemeString(lm), p.lit, LexemeString(p.k))
		return false
	}
	return true
}
