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

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ipointer := 0; ipointer < len(vm.instructions); ipointer++ {
		op := code.Opcode(vm.instructions[ipointer])

		switch op {
		case code.OpPop:
			vm.pop()
		case code.OpConstant:
			constPoolIdx := binary.BigEndian.Uint16(vm.instructions[ipointer+1:]) // get constpoolidx by decoding instructions
			ipointer += 2                                                         // increment the number of bytes

			if err := vm.push(vm.constants[constPoolIdx]); err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.executeBinOp(op); err != nil {
				return err
			}
		case code.OpTrue:
			if err := vm.push(&object.Boolean{Value: true}); err != nil {
				return err
			}
		case code.OpFalse:
			if err := vm.push(&object.Boolean{Value: false}); err != nil {
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

func (vm *VM) executeBinOp(op code.Opcode) error {
	r := vm.pop()
	l := vm.pop()

	if r.Type() == object.INTEGER_OBJ && l.Type() == object.INTEGER_OBJ {
		return vm.executeBinIntOp(op, r, l)
	}

	return fmt.Errorf("unsuported type for binop: %s, %s", r.Type(), l.Type())
}

func (vm *VM) executeBinIntOp(op code.Opcode, r, l object.Object) error {
	rValue := r.(*object.Integer).Value
	lValue := l.(*object.Integer).Value
	var res int64
	switch op {
	case code.OpAdd:
		res = lValue + rValue
	case code.OpSub:
		res = lValue - rValue
	case code.OpMul:
		res = lValue * rValue
	case code.OpDiv:
		res = lValue / rValue
	default:
		return fmt.Errorf("unknown integer op: %d", op)
	}
	return vm.push(&object.Integer{Value: res})
}
