package parser

import (
	"fmt"
	"strconv"

	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

const (
	_ int = iota // assign values 1 to 7 for the constants to get precedence
	LOWEST
	EQUALS      // ==
	LESSGREATER // > OR <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X OR !X
	CALL        // MyFunction(x)
	INDEX       // [1]
)

// precedence table to map token type to precedence
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	// only right side of exp
	prefixParseFn func() ast.Expression

	// left side of expression is passed as input
	infixParseFn func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	errors    []string
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(l *lexer.Lexer) (p *Parser) {
	p = &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.regPrefix(token.IDENT, p.parseIdentifier)
	p.regPrefix(token.INT, p.parseInteger)
	p.regPrefix(token.STRING, p.parseString)
	p.regPrefix(token.BANG, p.parsePrefixExpression)
	p.regPrefix(token.MINUS, p.parsePrefixExpression)
	p.regPrefix(token.TRUE, p.parseBoolean)
	p.regPrefix(token.FALSE, p.parseBoolean)
	p.regPrefix(token.LPAREN, p.parseGroupedExpressions)
	p.regPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.regPrefix(token.LBRACKET, p.parseArray)
	p.regPrefix(token.LBRACE, p.parseHashLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.regInfix(token.PLUS, p.parseInfixExpression)
	p.regInfix(token.MINUS, p.parseInfixExpression)
	p.regInfix(token.SLASH, p.parseInfixExpression)
	p.regInfix(token.ASTERISK, p.parseInfixExpression)
	p.regInfix(token.EQ, p.parseInfixExpression)
	p.regInfix(token.NOT_EQ, p.parseInfixExpression)
	p.regInfix(token.LT, p.parseInfixExpression)
	p.regInfix(token.GT, p.parseInfixExpression)
	p.regInfix(token.LPAREN, p.parseCallExpression)
	p.regInfix(token.LBRACKET, p.parseIndexExpression)

	p.nextToken() // initializes next token
	p.nextToken() // initializes curr token

	return
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) regPrefix(tt token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tt] = fn
}

func (p *Parser) regInfix(tt token.TokenType, fn infixParseFn) {
	p.infixParseFns[tt] = fn
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(tk token.TokenType) {
	err := fmt.Sprintf("expected next token to be %s, got %s instead", tk, p.peekToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		st := p.parseStatement()

		if st != nil {
			program.Statements = append(program.Statements, st)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	is := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	is.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	is.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		is.Alternative = p.parseBlockStatement()
	}

	return is
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	st := &ast.LetStatement{Token: p.curToken}

	// verify if token type is IDENTIFIER
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// create identifier node
	st.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	st.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return st
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	st := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	st.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}
	return st
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	// defer untrace(trace("parseExpressionStatement"))
	st := &ast.ExpressionStatement{Token: p.curToken}
	st.Expression = p.parseExpression(LOWEST) // we pass the lowest precedence operator because we didnt parse anything yet, so we cant compare precedence

	// we dont use expectPeek() because for expressions the semicolon is optional
	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return st
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		st := p.parseStatement()
		if st != nil {
			block.Statements = append(block.Statements, st)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseInteger() ast.Expression {
	// defer untrace(trace("parseIntegerLiteral"))
	intLiteral := &ast.IntegerLiteral{Token: p.curToken}

	val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as int", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	intLiteral.Value = val

	return intLiteral
}

func (p *Parser) parseArray() ast.Expression {
	a := &ast.ArrayLiteral{Token: p.curToken}
	l := []ast.Expression{}

	if p.peekToken.Type == token.RBRACKET {
		p.nextToken()
		return a
	}

	p.nextToken()
	l = append(l, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		l = append(l, p.parseExpression(LOWEST))
	}

	if p.peekToken.Type != token.RBRACKET {
		p.nextToken()
		l = nil
	}
	p.nextToken()
	a.Elements = l
	return a
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	e := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	e.Index = p.parseExpression(LOWEST)

	if p.peekToken.Type != token.RBRACKET {
		p.nextToken()
		return nil
	}
	p.nextToken()

	return e
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for p.peekToken.Type != token.RBRACE {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if p.peekToken.Type != token.COLON {
			p.nextToken()
			return nil
		}

		p.nextToken()
		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if p.peekToken.Type != token.RBRACE {
			if p.peekToken.Type != token.COMMA {
				p.nextToken()
				return nil
			}
			p.nextToken()
		}
	}

	if p.peekToken.Type != token.RBRACE {
		p.nextToken()
		return nil
	}

	p.nextToken()
	return hash
}

func (p *Parser) parseString() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	value := false
	if p.curToken.Type == token.TRUE {
		value = true
	}
	return &ast.Boolean{Token: p.curToken, Value: value}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	// creates a prefix operation node
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken() // advance the token pointer to get the integer or identifier

	// calls parseExpression to parse int or identifier and complete the node
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpressions() ast.Expression {
	p.nextToken()

	expression := p.parseExpression(LOWEST)

	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()
	return expression
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{
		Token: p.curToken,
	}

	if p.peekToken.Type != token.LPAREN {
		return nil
	}
	p.nextToken()

	fn.Arguments = p.parseFunctionArguments()

	if p.peekToken.Type != token.LBRACE {
		return nil
	}

	p.nextToken()

	fn.Body = p.parseBlockStatement()

	return fn
}

func (p *Parser) parseFunctionArguments() []*ast.Identifier {
	ids := []*ast.Identifier{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken() // advance token to rparen and return
		return ids
	}

	p.nextToken() // advance token to next id

	id := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	ids = append(ids, id)

	for p.peekToken.Type == token.COMMA {
		p.nextToken() // advance token to comma
		p.nextToken() // advance token to next id

		id := &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}

		ids = append(ids, id)
	}

	if p.peekToken.Type != token.RPAREN {
		return nil
	}
	p.nextToken()
	return ids
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.curToken, Function: function}
	call.Arguments = p.parseCallArguments()
	return call
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken()

	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if p.peekToken.Type != token.RPAREN {
		return nil
	}

	p.nextToken()
	return args
}

func (p *Parser) noPrefixParseError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	//defer untrace(trace("parseExpression"))
	/*
			   case 1 + 2 + 3
			       ast needs two infix expression nodes
			    a   first node:
			           - right node: 3
			           - operator: +
			           - left node: infix 1 + 2
			       second node:
			           - right node: 2
			           - operator: +
			           - left node: 1

			       tree:
			               ast.infixexpression
			               |                |
			       ast.infixexpression    ast.integer
			       |                 |         |
			   ast.integer     ast.integer     3
			       |                 |
			       1                 2

			   code working:
			       - parseExpression check if there is prefix function associated with curToken, and there is for 1 (INT);
			       - leftExp := *ast.IntegerLiteral
			       - for loop checks that peekToken is not semicolon and peekPrecedence is smaller
			       - inside for loop, fetch infixParse function assigned to next token (+);
			       - before executes, advances token so cur = + and peek = 2
			       - inside parse infix, creates ast.InfixExpression with operator = + and left = 1 (ast.IntegerLiteral) already defined
			       - saves the + precedence
			       - advances token
			       - calls parseExpression on the right part of the infix expression
		           - this call to parseExpression is done by parseInfixExpression, so when it returns, it returns to parseInfixExpression
			       - now parseExpression is called with cur = 2 and peek = +
			       - parse 2 as ast.IntegerLiteral
			       - for loop doesnt execute, because precedence of + (the argument passed as precedence) is equal than of peek (+), so ast.Integer 2 is returned
			       - it goes back to parseInfixExpression and the right node receives 2, constructing the ast.InfixExpression for 1 + 2
		           - parseInfixExpression returns to the outermost call of parseExpression, wich has precedence of LOWEST
			       - everything executes again, but left node is now ast.InfixExpression of 1 + 2 and right node is ast.IntegerLiteral of 3

	*/

	/*
	   case !!true

	   ast.Program
	       |
	   ast.ExpressionStatement
	       |
	   ast.prefixExpression(operator = !)
	       |
	   ast.prefixExpression(operator = !)
	       |
	   ast.BooleanLiteral
	       |
	      true

	   code:
	       - parseExpression checks if there is prefix for !, and there is (BANG);
	       - leftExp calls parsePrefixExpression
	       - creates prefix expression object and advances token
	       - cur = !, peek = true
	       - calls parseExpression for right
	       - leftExp calls parsePrefixExpression
	       - creates prefix expression objects and advances token
	       - cur = true, peek = nil
	       - calls parseExpression for right
	       - parse and returns boolean literal, because for loop doesnt enters
	       - returns expression with operator = ! and right = ast.Boolean(true)
	       - returns to outer call of parsePrefixExpression and creates new expression node
	       - returns to parseExpression and returns complete expression tree
	*/
	prefix := p.prefixParseFns[p.curToken.Type] // returns function associated with token type

	if prefix == nil {
		p.noPrefixParseError(p.curToken.Type)
		return nil
	}

	leftExpression := prefix() // calls function associated with token type and returns prefix expression in the form of ast.PrefixExpression

	for p.peekToken.Type != token.SEMICOLON && precedence < p.peekPrecedence() {
		// get infix function
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			p.noPrefixParseError(p.curToken.Type)
			return nil
		}
		p.nextToken() // advances token so that we can parse the new
		leftExpression = infix(leftExpression)
	}

	return leftExpression
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}
