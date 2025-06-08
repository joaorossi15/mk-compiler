package lexer

import (
	"monkey/token"
)

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

func New(s string) *Lexer {
	l := &Lexer{input: s}
	l.ReadChar()
	return l
}

func (l *Lexer) ReadChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}

	l.pos = l.readPos
	l.readPos += 1
}

func (l *Lexer) skipWhiteSpace() {
	for l.ch == '\t' || l.ch == '\r' || l.ch == ' ' || l.ch == '\n' {
		l.ReadChar()
	}
}

func (l *Lexer) isChar() bool {
	return l.ch >= 'a' && l.ch <= 'z' || l.ch >= 'A' && l.ch <= 'Z' || l.ch == '_'
}

func (l *Lexer) isDigit() bool {
	return l.ch >= '0' && l.ch <= '9'
}

func newToken(tk token.TokenType, s byte) token.Token {
	return token.Token{Type: tk, Literal: string(s)}
}

func (l *Lexer) createIdentifier() token.Token {
	var s string

	for l.isChar() {
		s += string(l.ch)
		l.ReadChar()
	}
	tp := token.LookupIdent(s)

	return token.Token{Type: tp, Literal: s}
}

func (l *Lexer) createInt() token.Token {
	var s string

	for l.isDigit() {
		s += string(l.ch)
		l.ReadChar()
	}

	return token.Token{Type: token.INT, Literal: s}
}

func (l *Lexer) readString() string {
	pos := l.pos + 1

	for {
		l.ReadChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}

	return l.input[pos:l.pos]
}

func (l *Lexer) NextToken() token.Token {
	var tk token.Token
	l.skipWhiteSpace()

	switch l.ch {
	case '+':
		tk = newToken(token.PLUS, l.ch)
	case '-':
		tk = newToken(token.MINUS, l.ch)
	case '*':
		tk = newToken(token.ASTERISK, l.ch)
	case '/':
		tk = newToken(token.SLASH, l.ch)
	case '(':
		tk = newToken(token.LPAREN, l.ch)
	case ')':
		tk = newToken(token.RPAREN, l.ch)
	case '{':
		tk = newToken(token.LBRACE, l.ch)
	case '}':
		tk = newToken(token.RBRACE, l.ch)
	case '[':
		tk = newToken(token.LBRACKET, l.ch)
	case ']':
		tk = newToken(token.RBRACKET, l.ch)
	case ';':
		tk = newToken(token.SEMICOLON, l.ch)
	case ':':
		tk = newToken(token.COLON, l.ch)
	case ',':
		tk = newToken(token.COMMA, l.ch)
	case '>':
		tk = newToken(token.GT, l.ch)
	case '<':
		tk = newToken(token.LT, l.ch)
	case '"':
		tk.Type = token.STRING
		tk.Literal = l.readString()
	case 0:
		tk.Literal = ""
		tk.Type = "EOF"
	case '!':
		if l.input[l.readPos] == '=' {
			l.ReadChar()
			tk = token.Token{Type: token.NOT_EQ, Literal: "!" + string(l.ch)}
		} else {
			tk = newToken(token.BANG, l.ch)
		}
	case '=':
		if l.input[l.readPos] == '=' {
			l.ReadChar()
			tk = token.Token{Type: token.EQ, Literal: string(l.ch) + string(l.ch)}
		} else {
			tk = newToken(token.ASSIGN, l.ch)
		}
	default:
		// parse identifiers: read new char until encounters a whitespace
		if l.isChar() {
			tk = l.createIdentifier()
			return tk
		} else if l.isDigit() {
			tk = l.createInt()
			return tk
		} else {
			tk = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.ReadChar()
	return tk
}
