package compiler

import (
	"fmt"
	"github.com/niksaak/rhmrm/asm/lexer"
	"github.com/niksaak/rhmrm/machine"
)

// Node of a parse tree.
type Node interface {
	Pos() lexer.Position
}

type ByteSpec struct {
	Offset   int  // offset in slice
	Size     uint // byte size in bits, zero means all
	Position int  // offset from words' lsb
}

type SymbolSpec struct {
	ByteSpec
	Name string
}

type (
	// ProgramNode represents the whole program.
	ProgramNode struct {
		lexer.Position
		Clauses []Node
	}

	// Toplevel node types //

	// LabelNode represents label definition.
	LabelNode struct {
		lexer.Position
		Name string
	}

	// DirectiveNode represents assembler directive.
	DirectiveNode struct {
		lexer.Position
		Op       string
		Operands []Node
	}

	// InstructionNode represents instruction.
	InstructionNode struct {
		lexer.Position
		Op       string
		Operands []Node
	}

	// CommentNode represents a comment.
	CommentNode struct {
		lexer.Position
		Level   int
		Comment string
	}

	// Operand node types //

	// RegisterNode represents a RHMRM register.
	RegisterNode struct {
		lexer.Position
		Kind  int
		Index int
	}

	// SymbolNode represents a symbol.
	SymbolNode struct {
		lexer.Position
		Name string
	}

	// IntegerNode represents an integer.
	IntegerNode struct {
		lexer.Position
		Value int
	}

	// StringNode represents a string.
	StringNode struct {
		lexer.Position
		Text string
	}

	// BlockNode represents a block of clauses.
	BlockNode struct {
		lexer.Position
		Clauses []Node
	}

	// TextNode is a slice of raw machine words.
	TextNode struct {
		lexer.Position
		Text []machine.Word
		Symbols []SymbolSpec
	}

	// ErrorNode represents parse error.
	ErrorNode struct {
		lexer.Position
		Message string
		Datum Node
	}
)

func (n ProgramNode) Pos() lexer.Position     { return n.Position }
func (n LabelNode) Pos() lexer.Position       { return n.Position }
func (n DirectiveNode) Pos() lexer.Position   { return n.Position }
func (n InstructionNode) Pos() lexer.Position { return n.Position }
func (n CommentNode) Pos() lexer.Position     { return n.Position }
func (n RegisterNode) Pos() lexer.Position    { return n.Position }
func (n SymbolNode) Pos() lexer.Position      { return n.Position }
func (n IntegerNode) Pos() lexer.Position     { return n.Position }
func (n StringNode) Pos() lexer.Position      { return n.Position }
func (n BlockNode) Pos() lexer.Position       { return n.Position }
func (n TextNode) Pos() lexer.Position        { return n.Position }
func (n ErrorNode) Pos() lexer.Position       { return n.Position }

// ErrorNode additionally implements error interface.
func (n ErrorNode) Error() (s string) {
	s = n.Message
	if n.Datum != nil {
		s += fmt.Sprintf("; DATUM: %v", n.Datum)
	}
	return s
}
