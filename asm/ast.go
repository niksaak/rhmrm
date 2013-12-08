package asm

import "fmt"

type Node interface {
	Pos() Position
	fmt.Stringer
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
		op SymbolNode
		operands []Node
	}

	// InstructionNode represents instruction.
	InstructionNode struct {
		DirectiveNode
	}

	// CommentNode represents a comment.
	CommentNode struct {
		Position
		level int
		comment string
	}

	// Operand node types //

	// RegisterNode represents a RHMRM register.
	RegisterNode struct {
		Position
		kind int
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
		ProgramNode
	}
)

func (n ProgramNode) Pos() Position { return n.Position }
func (n LabelNode) Pos() Position { return n.Position }
func (n DirectiveNode) Pos() Position { return n.Position }
func (n InstructionNode) Pos() Position { return n.Position }
func (n CommentNode) Pos() Position { return n.Position }
func (n RegisterNode) Pos() Position { return n.Position }
func (n SymbolNode) Pos() Position { return n.Position }
func (n IntegerNode) Pos() Position { return n.Position }
func (n StringNode) Pos() Position { return n.Position }
func (n BlockNode) Pos() Position { return n.Position }

func (n *ProgramNode) String() string {
	return fmt.Sprintf("program:( <%d clauses> )", len(n.clauses))
}

func (n *LabelNode) String() string {
	return fmt.Sprintf("label:%v", n.name)
}

// helper for DirectiveNode and InstructionNode
func cmdString(kind string, op Node, operands []Node) (s string) {
	s = fmt.Sprintf("%s:(%v", kind, op)
	for _, o := range operands {
		s += " " + o.String()
	}
	s += ")"
	return
}

func (n *DirectiveNode) String() string {
	return cmdString("directive", &n.op, n.operands)
}

func (n *InstructionNode) String() string {
	return cmdString("instruction", &n.op, n.operands)
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
