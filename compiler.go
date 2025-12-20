package main

import (
	"fmt"
	"strconv"
)

type Precedence int
type ParseFn func()

const DEBUG_PRINT_CODE = true

const (
	PREC_NONE       = iota
	PREC_ASSIGNMENT // =
	PREC_OR         // or
	PREC_AND        // and
	PREC_EQUALITY   // == !=
	PREC_COMPARISON // < > <= >=
	PREC_TERM       // + -
	PREC_FACTOR     // * /
	PREC_UNARY      // ! -
	PREC_CALL       // . ()
	PREC_PRIMARY
)

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

type Compiler struct {
	Sc    *Scanner
	Chunk *Chunk
	Ps    *Parser
	rules map[TokenType]ParseRule
}

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence Precedence
}

func (c *Compiler) compile(source string) bool {
	c.Sc.initScanner(source)
	c.initRules()
	c.advance()
	for !c.match(TOKEN_EOF) {
		c.declaration()
	}
	c.endCompiler()
	return !c.Ps.hadError
}

func (c *Compiler) declaration() {
	if c.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.statement()
	}

	if c.Ps.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) varDeclaration() {
	global := c.parseVariable("Expected variable name")
	if c.match(TOKEN_EQUAL) {
		c.expression()
	} else {
		c.emitByte(OP_NIL)
	}
	c.consume(TOKEN_SEMICOLON,
		"Expect ';' after variable declaration.")
	c.defineVariable(global)
}

func (c *Compiler) defineVariable(global byte) {
	c.emitBytes(OP_DEFINE_GLOBAL, global)
}

func (c *Compiler) parseVariable(errorMessage string) byte {
	c.consume(TOKEN_IDENTIFIER, errorMessage)
	return c.identifierConstant(c.Ps.previous)
}

func (c *Compiler) identifierConstant(name Token) byte {
	return c.makeConstant(ObjVal{Object: CreateStringObj(name.Lexeme)})
}

func (c *Compiler) synchronize() {
	c.Ps.panicMode = false
	for c.Ps.current.Type != TOKEN_EOF {
		if c.Ps.previous.Type == TOKEN_SEMICOLON {
			return
		}
		switch c.Ps.current.Type {
		case TOKEN_CLASS:
		case TOKEN_FUN:
		case TOKEN_VAR:
		case TOKEN_FOR:
		case TOKEN_IF:
		case TOKEN_WHILE:
		case TOKEN_PRINT:
		case TOKEN_RETURN:
			return
		default:
		}
		c.advance()
	}
}

func (c *Compiler) statement() {
	if c.match(TOKEN_PRINT) {
		c.printStatement()
	} else {
		c.expressionStatement()
	}
}

func (c *Compiler) match(tokType TokenType) bool {
	if !c.check(tokType) {
		return false
	}
	c.advance()
	return true
}

func (c *Compiler) check(tokType TokenType) bool {
	return c.Ps.current.Type == tokType
}

func (c *Compiler) printStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	c.emitByte(OP_PRINT)
}

func (c *Compiler) expressionStatement() {
	c.expression()
	c.consume(TOKEN_SEMICOLON, "Expect ';' after expression")
	c.emitByte(OP_POP)
}

func (c *Compiler) expression() {
	c.parsePrecedence(PREC_ASSIGNMENT)
}

func (c *Compiler) parsePrecedence(prec Precedence) {
	c.advance()
	prefixRule := c.getRule(c.Ps.previous.Type).prefix
	if prefixRule == nil {
		c.error("Expect expression.")
		return
	}
	prefixRule()
	for prec <= c.getRule(c.Ps.current.Type).precedence {
		c.advance()
		infixRule := c.getRule(c.Ps.previous.Type).infix
		infixRule()
	}
}

func (c *Compiler) number() {
	val, _ := strconv.ParseFloat(c.Ps.previous.Lexeme, 64)
	c.emitConstant(NumberVal(val))
}

func (c *Compiler) literal() {
	switch c.Ps.previous.Type {
	case TOKEN_NIL:
		c.emitByte(OP_NIL)
	case TOKEN_FALSE:
		c.emitByte(OP_FALSE)
	case TOKEN_TRUE:
		c.emitByte(OP_TRUE)
	}
}

func (c *Compiler) binary() {
	opType := c.Ps.previous.Type
	rule := c.getRule(opType)
	c.parsePrecedence(rule.precedence + 1)
	switch opType {
	case TOKEN_PLUS:
		c.emitByte(OP_ADD)
	case TOKEN_MINUS:
		c.emitByte(OP_SUBSTRACT)
	case TOKEN_STAR:
		c.emitByte(OP_MULTIPLY)
	case TOKEN_SLASH:
		c.emitByte(OP_DIVIDE)
	case TOKEN_GREATER:
		c.emitByte(OP_GREATER)
	case TOKEN_LESS:
		c.emitByte(OP_LESS)
	case TOKEN_EQUAL_EQUAL:
		c.emitByte(OP_EQUAL)
	case TOKEN_LESS_EQUAL:
		c.emitBytes(OP_GREATER, OP_NOT)
	case TOKEN_BANG_EQUAL:
		c.emitBytes(OP_EQUAL, OP_NOT)
	case TOKEN_GREATER_EQUAL:
		c.emitBytes(OP_LESS, OP_NOT)

	default:
		return // Unreachable.
	}
}

func (c *Compiler) str() {
	c.emitConstant(ObjVal{Object: CreateStringObj(c.Ps.previous.Lexeme[1 : len(c.Ps.previous.Lexeme)-1])})
}

func (c *Compiler) grouping() {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expected ')' after expression ")
}

func (c *Compiler) unary() {
	operatorType := c.Ps.previous.Type
	c.parsePrecedence(PREC_UNARY)
	switch operatorType {
	case TOKEN_MINUS:
		c.emitByte(OP_NEGATE)
	case TOKEN_BANG:
		c.emitByte(OP_NOT)
	default:
		return
	}
}

func (c *Compiler) variable() {
	c.namedVariable(c.Ps.previous)
}

func (c *Compiler) namedVariable(name Token) {
	arg := c.identifierConstant(name)
	c.emitBytes(OP_GET_GLOBAL, arg)
}

func (c *Compiler) emitConstant(val Value) {
	c.emitBytes(OP_CONSTANT, c.makeConstant(val))
}

func (c *Compiler) makeConstant(val Value) byte {
	constant := c.Chunk.AddConstant(val)
	if constant > 255 {
		c.error("Too many constants in one chunk")
		return 0
	}
	return byte(constant)

}

func (c *Compiler) emitByte(b byte) {
	c.Chunk.Write(b, c.Ps.previous.Line)
}

func (c *Compiler) endCompiler() {
	if DEBUG_PRINT_CODE {
		if !c.Ps.hadError {
			DisassembleChunk(c.Chunk, "code")
		}
	}
	c.emitReturn()
}

func (c *Compiler) emitReturn() {
	c.emitByte(OP_RETURN)
}

func (c *Compiler) emitBytes(b1, b2 byte) {
	c.emitByte(b1)
	c.emitByte(b2)
}

func (c *Compiler) consume(tokT TokenType, message string) {
	if c.Ps.current.Type == tokT {
		c.advance()
		return
	}

	c.errorAtCurrent(message)
}

func (c *Compiler) advance() {
	c.Ps.previous = c.Ps.current
	for {
		c.Ps.current = c.Sc.scanToken()
		if c.Ps.current.Type != TOKEN_ERROR {
			break
		}
		c.errorAtCurrent(c.Ps.current.Lexeme)
	}
}

func (c *Compiler) errorAtCurrent(message string) {
	c.errorAt(c.Ps.current, message)
}

func (c *Compiler) error(message string) {
	c.errorAt(c.Ps.previous, message)
}

func (c *Compiler) errorAt(tok Token, message string) {
	if c.Ps.panicMode {
		return
	}
	c.Ps.panicMode = true
	fmt.Printf("[line %d] Error", tok.Line)
	switch tok.Type {
	case TOKEN_EOF:
		fmt.Printf(" at end")
	case TOKEN_ERROR:

	default:
		fmt.Printf(" at '%s'", tok.Lexeme)
	}
	fmt.Printf(": %s\n", message)
	c.Ps.hadError = true
}

func (c *Compiler) initRules() {
	c.rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {c.grouping, nil, PREC_NONE},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, nil, PREC_NONE},
		TOKEN_MINUS:         {c.unary, c.binary, PREC_TERM},
		TOKEN_PLUS:          {nil, c.binary, PREC_TERM},
		TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		TOKEN_SLASH:         {nil, c.binary, PREC_FACTOR},
		TOKEN_STAR:          {nil, c.binary, PREC_FACTOR},
		TOKEN_BANG:          {c.unary, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, c.binary, PREC_EQUALITY},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, c.binary, PREC_EQUALITY},
		TOKEN_GREATER:       {nil, c.binary, PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {nil, c.binary, PREC_COMPARISON},
		TOKEN_LESS:          {nil, c.binary, PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {nil, c.binary, PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {c.variable, nil, PREC_NONE},
		TOKEN_STRING:        {c.str, nil, PREC_NONE},
		TOKEN_NUMBER:        {c.number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {c.literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {c.literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {c.literal, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func (c *Compiler) getRule(tok TokenType) ParseRule {
	return c.rules[tok]
}
