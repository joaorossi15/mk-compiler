package vm

import (
	"fmt"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

type vmTest struct {
	input    string
	expected interface{}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTest{
		{"true", true},
		{"false", false},
	}

	runVmTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTest{
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"2 / 1", 2},
	}

	runVmTests(t, tests)
}

func testIntegerObject(e int64, a object.Object) error {
	r, ok := a.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer, got %T", a)
	}

	if r.Value != e {
		return fmt.Errorf("wrong value, got %d, want %d", r.Value, e)
	}

	return nil
}

func testBooleanObject(e bool, a object.Object) error {
	r, ok := a.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean, got %T", a)
	}

	if r.Value != e {
		return fmt.Errorf("wrong value, got %t, want %t", r.Value, e)
	}

	return nil
}

func runVmTests(t *testing.T, tests []vmTest) {
	t.Helper()

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.NewParser(l)
		prog := p.ParseProgram()

		comp := compiler.New()
		if err := comp.Compile(prog); err != nil {
			t.Fatalf("compiler error %s", err)
		}

		vm := New(comp.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("vm error %s", err)
		}

		stackElement := vm.LastPoppedStackElement()
		testExpectedObj(t, tt.expected, stackElement)
	}
}

func testExpectedObj(t *testing.T, e interface{}, a object.Object) {
	t.Helper()

	switch e := e.(type) {
	case int:
		if err := testIntegerObject(int64(e), a); err != nil {
			t.Errorf("testing integer failed %s", err)
		}
	case bool:
		if err := testBooleanObject(bool(e), a); err != nil {
			t.Errorf("testing bool failed %s", err)
		}
	}
}
