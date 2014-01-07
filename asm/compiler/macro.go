package compiler

// genMacroExpander returns function accepting slice of len(operands) nodes
// and returning body with each symbol in operands replaced by corresponding
// node from args.
func (c *Compiler) genMacroExpander(
	operands []*SymbolNode,
	body []Node,
) TranslatorFunc {
	indices := make([][]int, len(operands)) // [operandIndex][]bodyPos
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
