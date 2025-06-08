package repl

import (
	"bufio"
	"fmt"
	"io"

	"monkey/eval"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
)

const MONKEY_FACE = ` 
     w  c(..)o   (
      \__(-)    __)
          /\   (
         /(_)___)
         w /|
          | \
         m  m
`

const PROMPT = `>> `

func CheckParser(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func REPL(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	env := object.NewEnviroment()

	for {
		fmt.Printf(">> ")

		scanned := scanner.Scan()

		if !scanned {
			if err := scanner.Err(); err != nil {
				return err
			}
			return nil
		}

		line := scanner.Text()

		if line == "exit" {
			return nil
		}

		l := lexer.New(line)
		p := parser.NewParser(l)
		prog := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printErrors(out, p.Errors())
			continue
		}

		evaluated := eval.Eval(prog, env)

		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}

	}
}

func printErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Looks like we ran into some monkey business here...\nparser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
