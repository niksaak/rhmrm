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
		ns = replace(ns, i, expanded...)
	}
	return ns
}

// generateText generates textNodes from instruction nodes.
func (c *Compiler) generateText(ns []Node) []Node {
	for i, n := range ns {
		n, ok := n.(*InstructionNode)
		if !ok {
			continue
		}
		// fn, ok := c.Instructions[n.Op]
		fn := c.InstructionMk(n.Op)
		if fn == nil {
			err := &ErrorNode{
				n.Pos(),
				"unknown instruction: " + n.Op,
			}
			c.ErrorCount++
			ns = replace(ns, i, err)
			continue
		}
		text := fn(n.Operands)
		// replace n with text in ns
		ns = replace(ns, i, text...)
	}
	return ns
}

// processSymbols resolves symbol references.
func (c *Compiler) processSymbols(ns []Node) []Node {
	ns = c.collectSymbols(ns)
	for i, nd := range ns {
		n, ok := nd.(*TextNode)
		if !ok {
			continue // skip non-text nodes
		}
		for _, r := range n.Symbols {
			v, ok := c.Symbols[r.Name]
			if !ok {
				err := &ErrorNode{
					n.Pos(),
					"unresolved symbol: " + r.Name,
				}
				c.ErrorCount++
				ns = replace(ns, i, err)
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
			err := &ErrorNode{n.Pos(), msg}
			return replace(ns, i, err)
		}
		ns = replace(ns, i, n)
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
		md, ok := nd.(*DirectiveNode)
		switch {
		case !ok: // macrodefs are declarations
			continue
		case md.Op != "macro": // macrodefs start with `macro`
			continue
		}
		if _, ok := md.Operands[0].(*SymbolNode); !ok {
			err := &ErrorNode{
				nd.Pos(),
				"malformed macro definition",
			}
			return replace(ns, i, err)
		}
		// TODO
	}
	return nil // TODO
}

// collectSymbols populates compiler state with symbol definitions.
func (c *Compiler) collectSymbols(ns []Node) []Node {
	return nil // TODO
}

// replace function replaces slice[n] with nodes.
func replace(slice []Node, n int, nodes ...Node) []Node {
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
