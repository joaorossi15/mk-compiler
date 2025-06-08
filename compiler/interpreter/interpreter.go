package interpreter

import (
	"fmt"
	"io"
	"os"

	"monkey/eval"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

func Interpreter(path string) error {
	env := object.NewEnviroment()
	var out io.Writer
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	l := lexer.New(string(f))
	p := parser.NewParser(l)
	prog := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printErrors(out, p.Errors())
		return err
	}

	evaluated := eval.Eval(prog, env)

	if evaluated != nil {
		fmt.Println(evaluated.Inspect())
	}

	return nil
}

func printErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Looks like we ran into some monkey business here...\nparser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, msg)
	}
}
