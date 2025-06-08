package ast

import (
	"bytes"
	"strings"

	"monkey/token"
)

// simple node interface
type Node interface {
	TokenLiteral() string
	String() string // method for printing and debugging
}

// node that represents a statemet (let, return, etc)
type Statement interface {
	Node
	statementNode()
}

// node that represents an expression (value)
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier // hold x in let x = 5;
	Value Expression  // expression that produces the value, 5 in let x = 5;
}

type ReturnStatement struct {
	Token token.Token // token.RETURN
	Value Expression  // expression that is returned
}

// statement that consists of only a single expression and works like a wrapper
// we need this because monkey is a script language, so lines with only a expressions statement are legal
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

// statement that happens inside a {} block
type BlockStatement struct {
	Token      token.Token // { token
	Statements []Statement
}

// hold x in let x = 5;
type Identifier struct {
	Token token.Token // name of the identifier, token.IDENT
	Value string      // value of the identifier, in this case x
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

type Boolean struct {
	Token token.Token // token.BOOL
	Value bool        // true or false
}

type StringLiteral struct {
	Token token.Token
	Value string
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

type IfStatement struct {
	Token       token.Token     // if token
	Condition   Expression      // condition for if to be executed
	Consequence *BlockStatement // { + code to be executed if passes
	Alternative *BlockStatement // { + code to be executed if doesnt passes
}

type FunctionLiteral struct {
	Token     token.Token     // fn token
	Arguments []*Identifier   // list containing all of the arguments
	Body      *BlockStatement // function body
}

type CallExpression struct {
	Token     token.Token // ( token
	Function  Expression  // identifier or function literal
	Arguments []Expression
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (lt *LetStatement) statementNode()       {}
func (lt *LetStatement) TokenLiteral() string { return lt.Token.Literal }

func (lt *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(lt.TokenLiteral() + " ")
	out.WriteString(lt.Name.String())
	out.WriteString(" = ")

	if lt.Value != nil {
		out.WriteString(lt.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (rt *ReturnStatement) statementNode()       {}
func (rt *ReturnStatement) TokenLiteral() string { return rt.Token.Literal }

func (rt *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rt.TokenLiteral())

	if rt.Value != nil {
		out.WriteString(rt.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// return whole expression as string
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (id *Identifier) expressionNode()      {}
func (id *Identifier) TokenLiteral() string { return id.Token.Literal }

func (id *Identifier) String() string { return id.Value }

func (bl *Boolean) expressionNode()      {}
func (bl *Boolean) TokenLiteral() string { return bl.Token.Literal }

func (bl *Boolean) String() string { return bl.Token.Literal }

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

func (il *IntegerLiteral) String() string { return il.Token.Literal }

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

func (sl *StringLiteral) String() string { return sl.Token.Literal }

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (i *IfStatement) statementNode()       {}
func (i *IfStatement) TokenLiteral() string { return i.Token.Literal }

func (i *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString(i.Token.Literal)

	out.WriteString(i.Condition.String())
	out.WriteString(" ")
	out.WriteString(i.Consequence.String())

	if i.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(i.Alternative.String())
	}

	return out.String()
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")

	for n, arg := range fl.Arguments {
		out.WriteString(arg.String())
		if n+1 == len(fl.Arguments) {
			break
		}
		out.WriteString(", ")
	}

	out.WriteString(")")
	out.WriteString(fl.Body.String())
	return out.String()
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

func (ce *CallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ce.Function.String())
	out.WriteString(ce.TokenLiteral())

	for n, arg := range ce.Arguments {
		out.WriteString(arg.String())
		if n+1 == len(ce.Arguments) {
			break
		}
		out.WriteString(", ")
	}

	out.WriteString(")")

	return out.String()
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	e := []string{}

	for _, el := range al.Elements {
		e = append(e, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(e, ", "))
	out.WriteString("]")

	return out.String()
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }

func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	p := []string{}

	for k, v := range hl.Pairs {
		p = append(p, k.String()+":"+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(p, ", "))
	out.WriteString("}")

	return out.String()
}

func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// returns the root of the program
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// program string method, writing the value of each satement String() method and returning it
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}
