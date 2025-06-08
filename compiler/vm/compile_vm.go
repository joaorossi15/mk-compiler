package vm

import (
	"fmt"
	"io"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/parser"
	"os"
)

func CompileRunVM(path string) error {
	var out io.Writer

	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	l := lexer.New(string(f))
	p := parser.NewParser(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printErrors(out, p.Errors())
		return err
	}

	c := compiler.New()

	if err := c.Compile(prog); err != nil {
		return err
	}

	virtualMachine := New(c.Bytecode())
	if err := virtualMachine.Run(); err != nil {
		return err
	}

	fmt.Printf("Stack: %s\n", virtualMachine.stack[:virtualMachine.sp])
	fmt.Println(virtualMachine.StackTop())
	return nil
}

func printErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Looks like we ran into some monkey business here...\nparser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, msg)
	}
}
