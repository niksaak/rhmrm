package main

import "fmt"

import (
	"github.com/niksaak/rhmrm/asm/compiler"
	"github.com/niksaak/rhmrm/asm/lexer"
	"github.com/niksaak/rhmrm/asm/parser"
	"github.com/niksaak/rhmrm/asm/util"
	"github.com/niksaak/rhmrm/machine"
)

var (
	_ compiler.Compiler
	_ lexer.Lexer
	_ parser.Parser
	_ = util.GeneralRegisterKind
	_ machine.Machine
)

func main() {
	fmt.Println("This executable is a linting stub")
	fmt.Println("if you can run it, you can build rhmrm packages.")
}
