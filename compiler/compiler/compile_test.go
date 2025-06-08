package compiler

import (
	"fmt"
	"monkey/code"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type compilerTests struct {
	input                string
	expectedInstructions []code.Instructions
	expectedConstants    []interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTests{
		{
			input: "1+2",
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
			},
			expectedConstants: []interface{}{"1", "2"},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTests) {
	t.Helper()

	for _, tt := range tests {
		lex := lexer.New(tt.input)
		p := parser.NewParser(lex)
		program := p.ParseProgram()
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(e []code.Instructions, a code.Instructions) error {
	out := code.Instructions{}
	for _, instruction := range e {
		out = append(out, instruction...)
	}

	if len(a) != len(out) {
		return fmt.Errorf("wrong i lenght. want %q, got %q", out, a)
	}

	for i, inst := range out {
		if a[i] != inst {
			return fmt.Errorf("wrong i at %d.\nwant %q\ngot %q", i, out, a)
		}
	}
	return nil
}

func testConstants(e []interface{}, a []object.Object) error {
	if len(a) != len(e) {
		return fmt.Errorf("wrong i lenght. want %d, got %d", len(a), len(e))
	}

	for i, c := range e {
		switch c := c.(type) {
		case int:
			if err := testIntegerObject(int64(c), a[i]); err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}
	return nil
}
