package asm

type Node interface {
	Pos() Position
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

// ErrorNode additionally implements error interface.
func (n ErrorNode) Error() string {
	return n.message
}
