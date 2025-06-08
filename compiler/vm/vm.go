package vm

import (
	"encoding/binary"
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	stack        []object.Object // expressions are objects in memory
	sp           int             // stack pointer, always points to the next value
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, 2048),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ipointer := 0; ipointer < len(vm.instructions); ipointer++ {
		op := code.Opcode(vm.instructions[ipointer])

		switch op {
		case code.OpConstant:
			constPoolIdx := binary.BigEndian.Uint16(vm.instructions[ipointer+1:]) // get constpoolidx by decoding instructions
			ipointer += 2                                                         // increment the number of bytes

			if err := vm.push(vm.constants[constPoolIdx]); err != nil {
				return err
			}
		case code.OpAdd:
			// get values from stack
			v1 := vm.pop()
			v2 := vm.pop()
			v1Value := v1.(*object.Integer).Value
			v2Value := v2.(*object.Integer).Value
			// add them
			if err := vm.push(&object.Integer{Value: (v1Value + v2Value)}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= 2048 {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}
