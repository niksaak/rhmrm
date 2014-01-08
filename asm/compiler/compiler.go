// Package compiler translates parse trees into machine code.
package compiler

import (
	"fmt"
	"github.com/niksaak/rhmrm/machine"
)

// TODO: parse tree traversal funcs are highly unoptimized. This is not cool.

// MacroFunc processes slice of operands into a slice of nodes.
type TranslatorFunc func(operands []Node) []Node

// Compiler implements translating an AST into machine code.
type Compiler struct {
	// Immutable state:
	Directives map[string]TranslatorFunc
	Mnemonics  map[string]machine.Word

	// Mutable state:
	Macros   map[string]TranslatorFunc
	Symbols  map[string]int
	Comments map[int]*CommentNode

	PassCount  int
	PassMax    int // when PassCount exceeds this, compiling is stopped
	ErrorCount int
}

// Init initializes compiler. Zero parameter sets passes cap to the default.
func (c *Compiler) Init(maxPasses int) *Compiler {
	c.Directives = make(map[string]TranslatorFunc)
	c.Mnemonics = make(map[string]machine.Word)
	c.Macros = make(map[string]TranslatorFunc)
	c.Symbols = make(map[string]int)
	c.Comments = make(map[int]*CommentNode)

	c.PassCount = 0
	if maxPasses != 0 {
		c.PassMax = maxPasses
	} else {
		c.PassMax = 32
	}
	c.ErrorCount = 0

	return c
}

// InstructionMk returns instruction
func (c *Compiler) InstructionMk(op string) TranslatorFunc {
	opcode, ok := c.Mnemonics[op]
	_ = opcode // TODO
	if !ok {
		return nil
	}
	return func(operands []Node) []Node {
		return nil // TODO
	}
}

func (c *Compiler) MacroMk(op string, operands []Node) TranslatorFunc {
	return nil // TODO
}

// Compile takes a parse tree and returns a slice of machine Words.
func (c *Compiler) Compile(nodes []Node) (_ []machine.Word, err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if _, ok := e.(error); !ok {
			err = fmt.Errorf("%v", e)
		}
	}()
	ret := make([]Node, len(nodes))
	copy(ret, nodes)
	ret = c.expandMacros(ret)
	ret = c.generateText(ret)
	ret = c.processSymbols(ret)
	return c.words(ret)
}

// expandMacros expands all macro invocations in source.
func (c *Compiler) expandMacros(ns []Node) []Node {
	ns = c.collectMacrodefs(ns)
	for i, n := range ns {
		n, ok := n.(*InstructionNode)
		if !ok { // node must be an instruction
			continue
		}
		fm, ok := c.Macros[n.Op]
		if !ok { // instruction operator must be a macro
			continue
		}
		expanded := c.expandMacros(fm(n.Operands))
		// replace n with expanded in ns
		ns = setNode(ns, i, expanded...)
	}
	return ns
}

// generateText generates textNodes from instruction nodes.
func (c *Compiler) generateText(ns []Node) []Node {
	for i, n := range ns {
		errorNd := mkErrorNodef(n)
		n, ok := n.(*InstructionNode)
		if !ok {
			continue
		}
		// fn, ok := c.Instructions[n.Op]
		fn := c.InstructionMk(n.Op)
		if fn == nil {
			c.ErrorCount++
			ns = setNode(ns, i,
				errorNd("unknown instruction: %s", n.Op))
			continue
		}
		text := fn(n.Operands)
		// replace n with text in ns
		ns = setNode(ns, i, text...)
	}
	return ns
}

// processSymbols resolves symbol references.
func (c *Compiler) processSymbols(ns []Node) []Node {
	ns = c.collectSymbols(ns)
	for i, nd := range ns {
		errorNd := mkErrorNodef(nd)
		n, ok := nd.(*TextNode)
		if !ok {
			continue // skip non-text nodes
		}
		for _, r := range n.Symbols {
			v, ok := c.Symbols[r.Name]
			if !ok {
				c.ErrorCount++
				ns = setNode(ns, i,
					errorNd("unresolved symbol: %s",
						r.Name))
				continue
			}
			setSpec(n.Text, r.ByteSpec, machine.Word(v))
			n.Symbols = n.Symbols[1:]
		}
		if len(n.Symbols) != 0 {
			msg := "unresolved symbols left: "
			for _, s := range n.Symbols {
				msg += s.Name + " "
			}
			ns = setNode(ns, i, errorNd(msg))
			c.ErrorCount++
			continue
		}
		ns = setNode(ns, i, n)
	}
	return ns
}

// words function concatenates a slice of TextNodes and returns slice of
// machine words. It returns error upon encountering any other kind of Node.
func (c *Compiler) words(ns []Node) ([]machine.Word, error) {
	ws := make([]machine.Word, 0, 255)
	for i, nd := range ns {
		text, ok := nd.(*TextNode)
		if !ok {
			err := fmt.Errorf("node #%d is not text-node in %v",
				i, ns)
			return ws, err
		}
		ws = append(ws, text.Text...)
	}
	return ws, nil
}

// collectMacrodefs populates compiler state with macro definitions.
func (c *Compiler) collectMacrodefs(ns []Node) []Node {
	for i, nd := range ns {
		errorNd := mkErrorNodef(nd)
		// macros are defined with the `.macro` directive.
		md, ok := nd.(*DirectiveNode)
		if !ok || md.Op != "macro" {
			continue
		}
		// first operand must be a symbol naming a macro
		ident, ok := md.Operands[0].(*SymbolNode)
		if !ok {
			ns = setNode(ns, i,
				errorNd("malformed macro definition"))
			c.ErrorCount++
			continue
		}
		// macro name must be unique
		name := ident.Name
		if _, ok := c.Symbols[name]; ok {
			ns = setNode(ns, i,
				errorNd("macro already defined: %s", name))
			c.ErrorCount++
			continue
		}
		// last operand of `.macro` is a block containing macro body
		body, ok := md.Operands[len(md.Operands)-1].(*BlockNode)
		if !ok {
			ns = setNode(ns, i, errorNd("missing macro body"))
			c.ErrorCount++
			continue
		}
		// other operands must be symbols
		args := make([]*SymbolNode, len(md.Operands)-2)
		for i, o := range md.Operands[1:len(md.Operands)-1] {
			if sym, ok := o.(*SymbolNode); ok {
				args[i] = sym
			} else {
				ns = setNode(ns, i,
					errorNd("not a symbol: %v", o))
				c.ErrorCount++
				continue
			}
		}
		// if all above is ok, compile a macro func
		fm := mkMacroExpander(args, body.Clauses)
		if fm == nil {
			ns = setNode(ns, i, errorNd("unable to compile macro"))
			continue
		}
		c.Macros[name] = fm
		ns = setNode(ns, i)
	}
	return ns
}

// collectSymbols populates compiler state with symbol definitions.
func (c *Compiler) collectSymbols(ns []Node) []Node {
	return nil // TODO
}

// mkMacroExpander returns function which takes slice of len(operands) nodes
// and returning body with each symbol in operands replaced by corresponding
// node from args.
func mkMacroExpander(
	operands []*SymbolNode,
	body []Node,
) TranslatorFunc {
	indices := make([][]int, len(operands)) // [operandIndex][_]bodyPos
	for i, n := range body {
		sym, ok := n.(*SymbolNode)
		if !ok {
			continue
		}
		for j, o := range operands {
			if sym.Name == o.Name {
				indices[j] = append(indices[j], i)
				break
			}
		}
	}
	return func(args []Node) []Node {
		if len(args) != len(operands) {
			panic("wrong number of arguments")
		}
		body := append([]Node{}, body...)
		for i, pos := range indices {
			for _, n := range pos {
				body[n] = args[i]
			}
		}
		return body
	}
}

// mkErrorNodef creates a function which takes format args and returns an
// ErrorNode pointer.
func mkErrorNodef(datum Node) func(string, ...interface{}) *ErrorNode {
	return func(msg string, args ...interface{}) *ErrorNode {
		return &ErrorNode{
			datum.Pos(),
			fmt.Sprintf(msg, args),
			datum,
		}
	}
}

// setNode function replaces slice[n] with nodes or, when no nodes
// are supplied, deletes slice[n] from slice.
func setNode(slice []Node, n int, nodes ...Node) []Node {
	if len(nodes) == 0 {
		copy(slice[:n], slice[:n+1])
		return slice[:len(slice)-1]
	}
	if len(nodes) == 1 { // no need to extend the slice in this case
		slice[n] = nodes[0]
		return slice
	} else {
		return append(slice[:n], append(nodes, slice[n+1:]...)...)
	}
}

// setSpec replaces range of bits in text defined by spec with those of val.
func setSpec(text []machine.Word, spec ByteSpec, val machine.Word) {
	mask := machine.Word(1<<spec.Size - 1)
	val &= mask
	text[spec.Offset] &^= mask
	text[spec.Offset] &= val << uint(spec.Position)
}
