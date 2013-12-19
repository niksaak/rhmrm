// Package compiler translates parse trees into machine code.
package compiler

import (
	"github.com/niksaak/rhmrm/asm/lexer"
	"github.com/niksaak/rhmrm/machine"
)

type ExpandableNode interface {
	Node
	Expand(*Compiler) Node
}

// TextNode contains translated machine words.
type TextNode struct {
	lexer.Position // position of the first node translated
	text []machine.Word
}

func (n *TextNode) Pos() lexer.Position { return n.Position }

// TranslatorFunc processes slice of operands into a slice of machine words.
type TranslatorFunc func(c *Compiler, operands []Node) []machine.Word

// MacroFunc processes slice of operands into a slice of nodes.
type MacroFunc func(c *Compiler, operands []Node) []Node

// Compiler implements translating an AST into machine code.
type Compiler struct {
	Directives   map[string]TranslatorFunc
	Macros       map[string]MacroFunc
	Instructions map[string]TranslatorFunc
	Symbols      map[string]Node
	unsolved     map[string]Node

	PassCount    int
	PassMax      int
	ErrorCount   int

	off          int // machine word offset for labels
	needPass     bool
}

func (c *Compiler) Init(maxPasses int) *Compiler {
	c.PassCount = 0
	if maxPasses != 0 {
		c.PassMax = maxPasses
	} else {
		c.PassMax = 32
	}
	c.ErrorCount = 0
	c.off = 0
	c.needPass = true

	return nil // TODO
}

// Translate takes a Node and returns the resulting machine code
// in a slice of words.
func (c *Compiler) Translate(tree Node) []machine.Word {
	return nil // TODO
}
