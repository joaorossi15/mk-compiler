package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
)

type Def struct {
	Name         string
	OperandBytes []int // number of bytes per operand
}

var definitions = map[Opcode]*Def{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},
	OpAdd:      {"OpAdd", []int{}},
	OpSub:      {"OpSub", []int{}},
	OpMul:      {"OpMul", []int{}},
	OpDiv:      {"OpDiv", []int{}},
	OpTrue:     {"OpTrue", []int{}},
	OpFalse:    {"OpFalse", []int{}},
}

func (is Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(is) {
		def, err := Lookup(is[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, is[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, is.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Def, operands []int) string {
	operandCount := len(def.OperandBytes)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func Lookup(op byte) (*Def, error) {
	def, ok := definitions[Opcode(op)]

	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// creates the instruction bytecode with op + operands
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instlen := 1
	for _, w := range def.OperandBytes {
		instlen += w
	}

	instructions := make([]byte, instlen)
	instructions[0] = byte(op)

	offset := 1
	for i, o := range operands {
		opWidth := def.OperandBytes[i]
		switch opWidth {
		case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))
		}
		offset += opWidth
	}

	return instructions
}

// translate the instruction bytecode
func ReadOperands(def *Def, is Instructions) ([]int, int) {
	op := make([]int, len(def.OperandBytes))
	offset := 0

	for i, w := range def.OperandBytes {
		switch w {
		case 2:
			op[i] = int(binary.BigEndian.Uint16(is[offset:]))
		}
		offset += w
	}

	return op, offset
}
