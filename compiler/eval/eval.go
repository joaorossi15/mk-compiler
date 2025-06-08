package eval

import (
	"fmt"
	"strings"

	"monkey/ast"
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"count": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				if substr, ok := args[1].(*object.String); ok {
					return &object.Integer{Value: int64(strings.Count(arg.Value, substr.Value))}
				} else {
					return newError("argument 1 to `count` not supported, got %s", args[1].Type())
				}
			default:
				return newError("argument 0 to `count` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}

			a := args[0].(*object.Array)
			if len(a.Elements) > 0 {
				return a.Elements[0]
			}
			return NULL
		},
	},
	"last": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
			}

			a := args[0].(*object.Array)
			if len(a.Elements) > 0 {
				return a.Elements[len(a.Elements)-1]
			}
			return NULL
		},
	},
	"tail": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `tail` must be ARRAY, got %s", args[0].Type())
			}

			a := args[0].(*object.Array)
			if len(a.Elements) > 0 {
				newElements := make([]object.Object, (len(a.Elements) - 1), (len(a.Elements) - 1))
				copy(newElements, a.Elements[1:(len(a.Elements))])
				return &object.Array{Elements: newElements}
			}
			return NULL
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

			a := args[0].(*object.Array)
			newElements := make([]object.Object, (len(a.Elements) + 1), (len(a.Elements) + 1))
			copy(newElements, a.Elements)
			newElements[len(a.Elements)] = args[1]
			return &object.Array{Elements: newElements}
		},
	},
}

var (
	TRUE       = &object.Boolean{Value: true}
	FALSE      = &object.Boolean{Value: false}
	NULL       = &object.Null{}
	OPERATIONS = map[string]func(int64, int64) int64{
		"+": func(a, b int64) int64 { return a + b },
		"-": func(a, b int64) int64 { return a - b },
		"*": func(a, b int64) int64 { return a * b },
		"/": func(a, b int64) int64 { return a / b },
	}
	BOOLOPERATIONS = map[string]func(int64, int64) bool{
		">":  func(a, b int64) bool { return a > b },
		"<":  func(a, b int64) bool { return a < b },
		"!=": func(a, b int64) bool { return a != b },
		"==": func(a, b int64) bool { return a == b },
	}
	STRCOMPOPERATIONS = map[string]func(string, string) bool{
		"!=": func(a, b string) bool { return a != b },
		"==": func(a, b string) bool { return a == b },
	}
)

func Eval(node ast.Node, env *object.Enviroment) object.Object {
	switch node := node.(type) {
	// statements
	case *ast.Program:
		return evalProgram(node, env) // calls evalStatements for all of the statements

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return &object.Return{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Add(node.Name.Value, val)

		// expressions
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.IfStatement:
		return evalIfStatement(node, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(left, node.Operator, right)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.FunctionLiteral:
		params := node.Arguments
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToObj(node.Value)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		e := evalExpressions(node.Elements, env)
		if len(e) == 1 && isError(e[0]) {
			return e[0]
		}
		return &object.Array{Elements: e}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		i := Eval(node.Index, env)
		if isError(i) {
			return i
		}
		return evalIndexExpression(left, i)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

func evalProgram(p *ast.Program, env *object.Enviroment) object.Object {
	var result object.Object

	for _, st := range p.Statements {
		result = Eval(st, env)

		switch result := result.(type) {
		case *object.Return:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Enviroment) object.Object {
	var result object.Object

	for _, st := range block.Statements {
		result = Eval(st, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Enviroment) object.Object {
	if val, ok := env.Value(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalExpressions(expressions []ast.Expression, env *object.Enviroment) []object.Object {
	var res []object.Object

	for _, e := range expressions {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		res = append(res, evaluated)
	}
	return res
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendedFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapedReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendedFunctionEnv(fn *object.Function, args []object.Object) *object.Enviroment {
	env := object.NewEnclosedEnviroment(fn.Env)
	for idx, param := range fn.Parameters {
		env.Add(param.Value, args[idx])
	}
	return env
}

func unwrapedReturnValue(obj object.Object) object.Object {
	if ret, ok := obj.(*object.Return); ok {
		return ret.Value
	}
	return obj
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(left, operator, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(left, operator, right)
	case operator == "==":
		return nativeBoolToObj(left == right)
	case operator == "!=":
		return nativeBoolToObj(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalIntegerInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	valueLeft := left.(*object.Integer).Value
	valueRight := right.(*object.Integer).Value
	if fn, ok := OPERATIONS[operator]; ok {
		return &object.Integer{Value: fn(valueLeft, valueRight)}
	} else if fn, ok := BOOLOPERATIONS[operator]; ok {
		return nativeBoolToObj(fn(valueLeft, valueRight))
	}

	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalStringInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	valueLeft := left.(*object.String).Value
	valueRight := right.(*object.String).Value
	if fn, ok := STRCOMPOPERATIONS[operator]; ok {
		return nativeBoolToObj(fn(valueLeft, valueRight))
	} else if operator == "+" {
		return &object.String{Value: valueLeft + valueRight}
	}

	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIfStatement(node *ast.IfStatement, env *object.Enviroment) object.Object {
	condition := Eval(node.Condition, env)

	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	} else {
		return NULL
	}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(arr, index object.Object) object.Object {
	a := arr.(*object.Array)
	i := index.(*object.Integer).Value

	maxIndex := int64(len(a.Elements) - 1)
	if i < 0 || i > maxIndex {
		return NULL
	}

	return a.Elements[i]
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Enviroment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range node.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return nil
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(v, env)
		if isError(value) {
			return nil
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	h := hash.(*object.Hash)
	k, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := h.Pairs[k.HashKey()]

	if !ok {
		return NULL
	}
	return pair.Value
}

func nativeBoolToObj(b bool) *object.Boolean {
	if b {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Value: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
