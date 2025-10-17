package main

import "fmt"

const (
	TOKEN_EOF        = "EOF"
	TOKEN_STRING     = "STRING"
	TOKEN_NUMBER     = "NUMBER"
	TOKEN_IDENTIFIER = "IDENTIFIER"

	TOKEN_EQUAL = "="
	TOKEN_PLUS  = "+"
	TOKEN_MINUS = "-"
	TOKEN_SLASH = "/"
	TOKEN_STAR  = "*"

	TOKEN_BANG          = "!"
	TOKEN_BANG_EQUAL    = "!="
	TOKEN_EQUAL_EQUAL   = "=="
	TOKEN_GREATER       = ">"
	TOKEN_GREATER_EQUAL = ">="
	TOKEN_LESS          = "<"
	TOKEN_LESS_EQUAL    = "<="

	TOKEN_COMMA     = ","
	TOKEN_SEMICOLON = ";"
	TOKEN_DOT       = "."

	TOKEN_LEFT_PAREN  = "("
	TOKEN_RIGHT_PAREN = ")"
	TOKEN_LEFT_BRACE  = "{"
	TOKEN_RIGHT_BRACE = "}"

	TOKEN_AND      = "AND"
	TOKEN_CLASS    = "CLASS"
	TOKEN_ELSE     = "ELSE"
	TOKEN_FALSE    = "FALSE"
	TOKEN_FUN      = "FUN"
	TOKEN_FOR      = "FOR"
	TOKEN_IF       = "IF"
	TOKEN_NIL      = "NIL"
	TOKEN_OR       = "OR"
	TOKEN_PRINT    = "PRINT"
	TOKEN_RETURN   = "RETURN"
	TOKEN_SUPER    = "SUPER"
	TOKEN_THIS     = "THIS"
	TOKEN_TRUE     = "TRUE"
	TOKEN_VAR      = "VAR"
	TOKEN_WHILE    = "WHILE"
	TOKEN_FUNCTION = "FUNCTION"
	TOKEN_LET      = "LET"
	TOKEN_ERROR    = "ERROR"
)

type TokenType string

var keywords = map[string]TokenType{
	"and":    TOKEN_AND,
	"class":  TOKEN_CLASS,
	"else":   TOKEN_ELSE,
	"if":     TOKEN_IF,
	"false":  TOKEN_FALSE,
	"true":   TOKEN_TRUE,
	"var":    TOKEN_VAR,
	"fun":    TOKEN_FUN,
	"while":  TOKEN_WHILE,
	"super":  TOKEN_SUPER,
	"print":  TOKEN_PRINT,
	"nil":    TOKEN_NIL,
	"or":     TOKEN_OR,
	"return": TOKEN_RETURN,
	"for":    TOKEN_FOR,
	"this":   TOKEN_THIS,
}

type Scanner struct {
	Source  string
	Start   int
	Current int
	Line    int
}

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
}

func (sc *Scanner) initScanner(source string) {
	sc.Source = source
	sc.Start = 0
	sc.Current = 0
	sc.Line = 1
}

func (sc *Scanner) scanToken() Token {
	sc.skipWhitespaces()
	sc.Start = sc.Current
	if sc.isAtEnd() {
		return sc.makeToken(TOKEN_EOF)
	}
	var tok TokenType
	c := sc.advance()
	if isDigit(c) {
		return sc.scanNumber()
	}
	switch c {
	case '(':
		return sc.makeToken(TOKEN_LEFT_PAREN)
	case ')':
		return sc.makeToken(TOKEN_RIGHT_PAREN)
	case '{':
		return sc.makeToken(TOKEN_LEFT_BRACE)
	case '}':
		return sc.makeToken(TOKEN_RIGHT_BRACE)
	case ';':
		return sc.makeToken(TOKEN_SEMICOLON)
	case ',':
		return sc.makeToken(TOKEN_COMMA)
	case '.':
		return sc.makeToken(TOKEN_DOT)
	case '-':
		return sc.makeToken(TOKEN_MINUS)
	case '+':
		return sc.makeToken(TOKEN_PLUS)
	case '/':
		return sc.makeToken(TOKEN_SLASH)
	case '*':
		return sc.makeToken(TOKEN_STAR)
	case '!':
		tok = TOKEN_BANG
		if sc.match('=') {
			tok = TOKEN_BANG_EQUAL
		}
		return sc.makeToken(tok)
	case '=':
		tok = TOKEN_EQUAL
		if sc.match('=') {
			tok = TOKEN_EQUAL_EQUAL
		}
		return sc.makeToken(tok)
	case '<':
		tok = TOKEN_LESS
		if sc.match('=') {
			tok = TOKEN_LESS_EQUAL
		}
		return sc.makeToken(tok)
	case '>':
		tok = TOKEN_GREATER
		if sc.match('=') {
			tok = TOKEN_GREATER_EQUAL
		}
		return sc.makeToken(tok)
	case '"':
		return sc.scanString()
	}
	return sc.errorToken(fmt.Sprintf("Unexpected character %s", string(c)))

}

func (sc *Scanner) scanString() Token {
	for !sc.isAtEnd() && sc.getCharAtPos(sc.Current) != '"' {
		if sc.getCharAtPos(sc.Current) == '\n' {
			sc.Line++
		}
		sc.advance()
	}
	if sc.isAtEnd() {
		return sc.errorToken("Uterminated string")
	}
	sc.advance()
	return sc.makeToken(TOKEN_STRING)
}

func (sc *Scanner) scanNumber() Token {
	for isDigit(sc.getCharAtPos(sc.Current)) {
		sc.advance()
	}

	if sc.getCharAtPos(sc.Current) == '.' && isDigit(sc.getCharAtPos(sc.Current+1)) {
		sc.advance()
		for isDigit(sc.getCharAtPos(sc.Current)) {
			sc.advance()
		}
	}

	return sc.makeToken(TOKEN_NUMBER)
}

func (sc *Scanner) scanIdentifier() Token {
	for isAlpha(sc.getCharAtPos(sc.Current)) || isDigit(sc.getCharAtPos(sc.Current)) {
		sc.advance()
	}
	iden := sc.Source[sc.Start:sc.Current]
	var tok TokenType = TOKEN_IDENTIFIER
	if keyword, ok := keywords[iden]; ok {
		tok = keyword
	}
	return sc.makeToken(tok)
}

func (sc *Scanner) isAtEnd() bool {
	return sc.Current == len(sc.Source)
}

func (sc *Scanner) makeToken(tokenType TokenType) Token {
	return Token{
		Type:   tokenType,
		Lexeme: sc.Source[sc.Start:sc.Current],
		Line:   sc.Line,
	}
}

func (sc *Scanner) errorToken(message string) Token {
	return Token{
		Type:   TOKEN_ERROR,
		Lexeme: message,
		Line:   sc.Line,
	}
}

func (sc *Scanner) advance() int32 {
	sc.Current++
	return sc.getCharAtPos(sc.Current - 1)
}

func (sc *Scanner) match(expected int32) bool {
	if sc.isAtEnd() {
		return false
	}
	if sc.getCharAtPos(sc.Current) != expected {
		return false
	}
	sc.Current++
	return true

}

func (sc *Scanner) skipWhitespaces() {
	c := sc.getCharAtPos(sc.Current)
	switch c {
	case ' ', '\t', 'r':
		sc.advance()
	case '\n':
		sc.Line++
		sc.advance()
	case '/':
		if sc.getCharAtPos(sc.Current+1) == '/' {
			for sc.getCharAtPos(sc.Current) != '\n' {
				sc.advance()
			}
		}
	default:
		return
	}
}

func (sc *Scanner) getCharAtPos(pos int) int32 {
	if pos > len(sc.Source)-1 {
		return '\x00'
	}
	return []rune(sc.Source)[pos]
}

func isDigit(c int32) bool {
	return c > '0' && c < '9'
}

func isAlpha(c int32) bool {
	return (c > 'a' && c < 'z') || (c > 'A' && c < 'Z') || c == '_'
}
