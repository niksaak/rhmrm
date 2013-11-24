package asm

import "fmt"

var indent_level int
var indent_size int = 2

func indent() int {
	return indent_level * indent_size
}

// Node kinds.
const (
	ASTKind = iota

	// Toplevel kinds
	LabelKind
	InstructionKind
	DirectiveKind
	CommentKind

	// Subtree block
	BlockKind

	// Operators
	InstructionOpKind
	DirectiveOpKind

	// Operands
	StringKind
	SymbolKind
	RegisterKind
	IntegerKind
)

// Node of a syntax tree.
type Node interface {
	Kind() int
	Pos() Position
	fmt.Stringer
}

type AST struct {
	file string
	nodes []Node
}

func (a AST) Kind() int { return ASTKind }

func (a AST) Pos() Position { return Position{ a.file, 0, 1, 1 } }

func (a AST) String() (s string) {
	s = fmt.Sprintf("%*s:(\n", indent(), a.file)
	indent_level++
	for _, n := range a.nodes {
		s += fmt.Sprintf("%*s\n", indent(), n)
	}
	indent_level--
	s += ")\n"
	return
}

// LabelNode represents label definition in syntax tree.
type LabelNode struct {
	pos Position
	name string
}

func (l LabelNode) Kind() int { return LabelKind }

func (l LabelNode) Pos() Position { return l.pos }

func (l LabelNode) String() string {
	return fmt.Sprintf("%*s", indent(), l.name)
}

// InstructionNode represents instruction statement.
type InstructionNode struct {
	pos Position
	data []Node
}

func (i InstructionNode) Kind() int { return InstructionKind }

func (i InstructionNode) Pos() Position { return i.pos }

func (i InstructionNode) String() (s string) {
	s = fmt.Sprintf("%*s", indent(), "instruction:(")
	for n := range i.data {
		s += i.data[n].String()
		if n < len(i.data) - 1 {
			s += " "
		} else {
			s += ")"
		}
	}
	return
}

// DirectiveNode represents assembler directives.
type DirectiveNode struct {
	pos Position
	data []Node
}

func (d DirectiveNode) Kind() int { return DirectiveKind }

func (d DirectiveNode) Pos() Position { return d.pos }

func (d DirectiveNode) String() (s string) {
	s = fmt.Sprintf("%*s", indent(), "directive:(")
	for n := range d.data {
		s += d.data[n].String()
		if n < len(d.data) - 1 {
			s += " "
		} else {
			s += ")"
		}
	}
	return
}

// CommentNode represents source comment.
type CommentNode struct {
	pos Position
	comment string
}

func (c CommentNode) Kind() int { return CommentKind }

func (c CommentNode) Pos() Position { return c.pos }

func (c CommentNode) String() string {
	return fmt.Sprintf("%*s:%q", indent(), "comment", c.comment)
}

// BlockNode represents block.
type BlockNode struct {
	pos Position
	nodes []Node
}

func (b BlockNode) Kind() int { return BlockKind }

func (b BlockNode) Pos() Position { return b.pos }

func (b BlockNode) String() (s string) {
	s = fmt.Sprintf("%*s:(\n", indent(), "block")
	indent_level++
	for _, n := range b.nodes {
		s += fmt.Sprintf("%*s\n", indent(), n)
	}
	indent_level--
	s += ")\n"
	return
}

// InstructionOpNode represents instruction operator.
type InstructionOpNode struct {
	pos Position
	name string
}

func (o InstructionOpNode) Kind() int { return InstructionOpKind }

func (o InstructionOpNode) Pos() Position { return o.pos }

func (o InstructionOpNode) String() string {
	return "op:" + o.name
}

// DirectiveOpNode represents directive operator.
type DirectiveOpNode struct {
	pos Position
	name string
}

func (o DirectiveOpNode) Kind() int { return DirectiveOpKind }

func (o DirectiveOpNode) Pos() Position { return o.pos }

func (o DirectiveOpNode) String() string {
	return "dop:" + o.name
}

// StringNode represents string.
type StringNode struct {
	pos Position
	str string
}

func (s StringNode) Kind() int { return StringKind }

func (s StringNode) Pos() Position { return s.pos }

func (s StringNode) String() string {
	return fmt.Sprintf("string:%q", s.str)
}

// SymbolNode represents a symbol.
type SymbolNode struct {
	pos Position
	name string
}

func (s SymbolNode) Kind() int { return SymbolKind }

func (s SymbolNode) Pos() Position { return s.pos }

func (s SymbolNode) String() string {
	return "symbol:" + s.name
}

// RegisterNode represents register name.
type RegisterNode struct {
	pos Position
	name string
}

func (r RegisterNode) Kind() int { return RegisterKind }

func (r RegisterNode) Pos() Position { return r.pos }

func (r RegisterNode) String() string {
	return "register:" + r.name
}

// IntegerNode represents an integer.
type IntegerNode struct {
	pos Position
	num int
}

func (i IntegerNode) Kind() int { return IntegerKind }

func (i IntegerNode) Pos() Position { return i.pos }

func (i IntegerNode) String() string {
	return fmt.Sprintf("int:%d", i.num)
}
