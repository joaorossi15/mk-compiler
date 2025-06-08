package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"monkey/ast"
)

type ObjectType string

type BuiltinFunction func(args ...Object) Object

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ERROR_OBJ        = "ERROR"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
	ARRAY_OBJ        = "ARRAY"
	NULL_OBJ         = "NULL"
	HASH_OBJ         = "HASH"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value int64
}

type Boolean struct {
	Value bool
}

type String struct {
	Value string
}

type Return struct {
	Value Object
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Enviroment
}

type Builtin struct {
	Fn BuiltinFunction
}

type Array struct {
	Elements []Object
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object // to be able to print keys and values and not hashkey and values
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

type Error struct {
	Value string
}

type Null struct{}

type Enviroment struct {
	store map[string]Object
	outer *Enviroment
}

func (b *Boolean) HashKey() HashKey {
	var v uint64

	if b.Value {
		v = 1
	} else {
		v = 0
	}

	return HashKey{Type: b.Type(), Value: v}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func NewEnclosedEnviroment(outer *Enviroment) *Enviroment {
	env := NewEnviroment()
	env.outer = outer
	return env
}

func NewEnviroment() *Enviroment {
	s := make(map[string]Object)
	return &Enviroment{store: s, outer: nil}
}

func (e *Enviroment) Value(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Value(name)
	}
	return obj, ok
}

func (e *Enviroment) Add(name string, obj Object) Object {
	e.store[name] = obj
	return obj
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

func (rt *Return) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rt *Return) Inspect() string  { return rt.Value.Inspect() }

func (bt *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (bt *Builtin) Inspect() string  { return "builtin function" }

func (ar *Array) Type() ObjectType { return ARRAY_OBJ }
func (ar *Array) Inspect() string {
	var out bytes.Buffer

	e := []string{}

	for _, el := range ar.Elements {
		e = append(e, el.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(e, ", "))
	out.WriteString("]")
	return out.String()
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}

	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (err *Error) Type() ObjectType { return ERROR_OBJ }
func (err *Error) Inspect() string  { return "ERROR: " + err.Value }

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
