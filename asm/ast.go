package asm

import (
	"fmt"
	"io"
)

type Node interface {
	Pos() Position
	fmt.Stringer
}

type TreeNode interface {
	Node
	Name() string
	Tree() []Node
}

// Register kinds.
const (
	generalRegisterKind = iota << 28
	controlRegisterKind
	// to be expanded in future
)

// Control register acces modes.
const (
	controlSETMode = iota
	controlANDMode
	controlIORMode
	controlXORMode
)

var controlModes = map[rune]int{
	'=': controlSETMode,
	'&': controlANDMode,
	'|': controlIORMode,
	'^': controlXORMode,
}

type (
	// ProgramNode represents the whole program.
	ProgramNode struct {
		Position
		clauses []Node
	}

	// Toplevel node types //

	// LabelNode represents label definition.
	LabelNode struct {
		Position
		name string
	}

	// DirectiveNode represents assembler directive.
	DirectiveNode struct {
		Position
		op       string
		operands []Node
	}

	// InstructionNode represents instruction.
	InstructionNode struct {
		Position
		op       string
		operands []Node
	}

	// CommentNode represents a comment.
	CommentNode struct {
		Position
		level   int
		comment string
	}

	// Operand node types //

	// RegisterNode represents a RHMRM register.
	RegisterNode struct {
		Position
		kind  int
		index int
	}

	// SymbolNode represents a symbol.
	SymbolNode struct {
		Position
		name string
	}

	// IntegerNode represents an integer.
	IntegerNode struct {
		Position
		value int
	}

	// StringNode represents a string.
	StringNode struct {
		Position
		text string
	}

	// BlockNode represents a block of clauses.
	BlockNode struct {
		Position
		clauses []Node
	}

	// ErrorNode represents parse error.
	ErrorNode struct {
		Position
		message string
	}
)

func (n ProgramNode) Pos() Position     { return n.Position }
func (n LabelNode) Pos() Position       { return n.Position }
func (n DirectiveNode) Pos() Position   { return n.Position }
func (n InstructionNode) Pos() Position { return n.Position }
func (n CommentNode) Pos() Position     { return n.Position }
func (n RegisterNode) Pos() Position    { return n.Position }
func (n SymbolNode) Pos() Position      { return n.Position }
func (n IntegerNode) Pos() Position     { return n.Position }
func (n StringNode) Pos() Position      { return n.Position }
func (n BlockNode) Pos() Position       { return n.Position }
func (n ErrorNode) Pos() Position       { return n.Position }

func (n *ProgramNode) String() string {
	return fmt.Sprintf("program:( <%d clauses> )", len(n.clauses))
}

func (n *LabelNode) String() string {
	return fmt.Sprintf("label:%v", n.name)
}

// helper for DirectiveNode and InstructionNode
func cmdString(kind string, op string, operands []Node) (s string) {
	s = fmt.Sprintf("%s:(%s", kind, op)
	for _, o := range operands {
		s += " " + o.String()
	}
	s += ")"
	return
}

func (n *DirectiveNode) String() string {
	return cmdString("directive", n.op, n.operands)
}

func (n *InstructionNode) String() string {
	return cmdString("instruction", n.op, n.operands)
}

func (n *CommentNode) String() string {
	return fmt.Sprintf("comment:(%d %q)", n.level, n.comment)
}

func (n *RegisterNode) String() string {
	var kind string
	switch n.kind >> 28 {
	case generalRegisterKind:
		kind = "r"
	case controlRegisterKind:
		switch n.kind &^ 0xf000 {
		case controlANDMode:
			kind = "&"
		case controlIORMode:
			kind = "|"
		case controlXORMode:
			kind = "^"
		}
		kind += "c"
	default:
		kind = "X"
	}
	return fmt.Sprintf("register:%s%d", kind, n.index)
}

func (n *SymbolNode) String() string {
	return "symbol:" + n.name
}

func (n *IntegerNode) String() string {
	return fmt.Sprintf("integer:%d", n.value)
}

func (n *StringNode) String() string {
	return fmt.Sprintf("string:%q", n.text)
}

func (n *BlockNode) String() string {
	return fmt.Sprintf("block:( <%d clauses> )", len(n.clauses))
}

func (n *ErrorNode) String() string {
	return fmt.Sprintf("ERROR:%q", n.message)
}

func (n *ProgramNode) Name() string { return "program" }
func (n *BlockNode) Name() string   { return "block" }

func (n *ProgramNode) Tree() []Node { return n.clauses }
func (n *BlockNode) Tree() []Node   { return n.clauses }

// ErrorNode additionally implements error interface.
func (n ErrorNode) Error() string {
	return n.message
}

// PrintAST outputs abstract syntax tree into an io.Writer.
func PrintAST(node Node, w io.Writer) (err error) {
	putf := func(format string, args ...interface{}) {
		_, err = fmt.Fprintf(w, format, args...)
	}
	var putnd func(Node, string)
	putnd = func(n Node, indent string) {
		if t, ok := n.(TreeNode); ok {
			putf(indent + t.Name() + ":(")
			if err != nil {
				return
			}
			for _, nd := range t.Tree() {
				putf("\n")
				putnd(nd, indent+"  ")
				if err != nil {
					return
				}
			}
			putf("\n" + indent + ")")
		} else {
			putf(indent+"%v", n)
		}
	}

	putnd(node, "")
	if err != nil {
		return
	}
	putf("\n")

	return
}
