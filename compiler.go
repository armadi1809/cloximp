package main

import (
	"fmt"
	"math"
	"strconv"
)

type Precedence int
type FunctionType int
type ParseFn func(canAssign bool)

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

const (
	TYPE_FUNCTION = iota
	TYPE_SCRIPT
)

type Parser struct {
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
}

type Local struct {
	name  Token
	depth int
}

type Compiler struct {
	Sc         *Scanner
	Function   *ObjFunction
	Type       FunctionType
	Ps         *Parser
	rules      map[TokenType]ParseRule
	Locals     []Local
	LocalCount int
	ScopeDepth int
}

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence Precedence
}

func (c *Compiler) compile(source string) *ObjFunction {
	c.Sc.initScanner(source)
	c.initRules()
	c.advance()
	for !c.match(TOKEN_EOF) {
		c.declaration()
	}
	function := c.endCompiler()

	if c.Ps.hadError {
		return nil
	}

	return function
}

func (c *Compiler) declaration() {
	if c.match(TOKEN_FUN) {
		c.functionDeclaration()
	} else if c.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.statement()
	}

	if c.Ps.panicMode {
		c.synchronize()
	}
}

func (c *Compiler) functionDeclaration() {
	global := c.parseVariable("Expect function name.")
	c.markInitialized()
	c.defineVariable(global)
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
	if c.ScopeDepth > 0 {
		c.markInitialized()
		return
	}
	c.emitBytes(OP_DEFINE_GLOBAL, global)
}

func (c *Compiler) markInitialized() {
	if c.ScopeDepth == 0 {
		return
	}
	c.Locals[c.LocalCount-1].depth =
		c.ScopeDepth
}

func (c *Compiler) declareVariable() {
	if c.ScopeDepth == 0 {
		return
	}
	name := c.Ps.previous
	for i := c.LocalCount; i >= 0; i-- {
		local := c.Locals[i]
		if local.depth != -1 && local.depth < c.ScopeDepth {
			break
		}
		if identifiersEqual(name, local.name) {
			c.error("Already a variable with this name in this scope.")
		}
	}
	c.addLocal(name)

}

func (c *Compiler) addLocal(name Token) {
	if c.LocalCount == 255 {
		c.error("Too many local variables in function.")
		return
	}
	local := Local{}
	local.depth = -1
	local.name = name
	c.Locals = append(c.Locals, local)
}

func (c *Compiler) parseVariable(errorMessage string) byte {
	c.consume(TOKEN_IDENTIFIER, errorMessage)
	c.declareVariable()
	if c.ScopeDepth > 0 {
		return 0
	}
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
	} else if c.match(TOKEN_IF) {
		c.ifStatement()
	} else if c.match(TOKEN_WHILE) {
		c.whileStatement()
	} else if c.match(TOKEN_FOR) {
		c.forStatement()
	} else if c.match(TOKEN_LEFT_BRACE) {
		c.beginBlock()
		c.block()
		c.endBlock()
	} else {
		c.expressionStatement()
	}
}

func (c *Compiler) forStatement() {
	c.beginBlock()
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'for'.")
	if c.match(TOKEN_SEMICOLON) {
		// No init, just keep going
	} else if c.match(TOKEN_VAR) {
		c.varDeclaration()
	} else {
		c.expressionStatement()
	}

	loopStart := c.Function.chunk.Count()
	exitJump := -1
	if !c.match(TOKEN_SEMICOLON) {
		c.expression()
		c.consume(TOKEN_SEMICOLON, "Expect ';' after loop condition.")

		exitJump = c.emitJump(OP_JUMP_IF_FALSE)
		c.emitByte(OP_POP)

	}
	if !c.match(TOKEN_RIGHT_PAREN) {
		bodyJump := c.emitJump(OP_JUMP)
		incrementStart := c.Function.chunk.Count()
		c.expression()
		c.emitByte(OP_POP)
		c.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")

		c.emitLoop(loopStart)
		loopStart = incrementStart
		c.patchJump(bodyJump)
	}

	c.statement()
	c.emitLoop(loopStart)

	if exitJump != -1 {
		c.patchJump(exitJump)
		c.emitByte(OP_POP) // Condition.
	}
	c.endBlock()
}

func (c *Compiler) whileStatement() {
	loopStart := c.Function.chunk.Count()
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'while'.")
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect '(' after condition.")

	exitJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.emitByte(OP_POP)
	c.statement()
	c.emitLoop(loopStart)
	c.patchJump(exitJump)
	c.emitByte(OP_POP)
}

func (c *Compiler) emitLoop(loopStart int) {
	c.emitByte(OP_LOOP)
	offset := c.Function.chunk.Count() - loopStart + 2
	if offset > math.MaxUint16 {
		c.error("Loop body too large.")
	}
	c.emitByte(byte((offset >> 8) & 0xff))
	c.emitByte(byte(offset & 0xff))

}

func (c *Compiler) ifStatement() {
	c.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expect '(' after condition.")

	thenJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.emitByte(OP_POP)

	c.statement()
	elseJump := c.emitJump(OP_JUMP)

	c.patchJump(thenJump)
	c.emitByte(OP_POP)
	if c.match(TOKEN_ELSE) {
		c.statement()
	}
	c.patchJump(elseJump)
}

func (c *Compiler) patchJump(offset int) {
	jump := c.Function.chunk.Count() - offset - 2
	if jump > math.MaxUint16 {
		c.error("Too many instructions to jump over")
	}
	c.Function.chunk.Code[offset] = byte((jump >> 8) & 0xff)
	c.Function.chunk.Code[offset+1] = byte(jump & 0xff)
}

func (c *Compiler) emitJump(instruction byte) int {
	c.emitByte(instruction)
	c.emitByte(0xff)
	c.emitByte(0xff)

	return c.Function.chunk.Count() - 2
}

func (c *Compiler) block() {
	for !c.check(TOKEN_RIGHT_BRACE) && !c.check(TOKEN_EOF) {
		c.declaration()
	}
	c.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (c *Compiler) beginBlock() {
	c.ScopeDepth += 1
}

func (c *Compiler) endBlock() {
	c.ScopeDepth -= 1
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
	canAssign := prec <= PREC_ASSIGNMENT
	prefixRule(canAssign)
	for prec <= c.getRule(c.Ps.current.Type).precedence {
		c.advance()
		infixRule := c.getRule(c.Ps.previous.Type).infix
		infixRule(canAssign)
	}

	if canAssign && c.match(TOKEN_EQUAL) {
		c.error("Invalid assignment target.")
	}
}

func (c *Compiler) number(canAssign bool) {
	val, _ := strconv.ParseFloat(c.Ps.previous.Lexeme, 64)
	c.emitConstant(NumberVal(val))
}

func (c *Compiler) literal(canAssign bool) {
	switch c.Ps.previous.Type {
	case TOKEN_NIL:
		c.emitByte(OP_NIL)
	case TOKEN_FALSE:
		c.emitByte(OP_FALSE)
	case TOKEN_TRUE:
		c.emitByte(OP_TRUE)
	}
}

func (c *Compiler) and_(canAssign bool) {
	endJump := c.emitJump(OP_JUMP_IF_FALSE)
	c.emitByte(OP_POP)
	c.parsePrecedence(PREC_AND)

	c.patchJump(endJump)
}

func (c *Compiler) or_(canAssign bool) {
	elseJump := c.emitJump(OP_JUMP_IF_FALSE)
	endJump := c.emitJump(OP_JUMP)

	c.patchJump(elseJump)
	c.emitByte(OP_POP)

	c.parsePrecedence(PREC_OR)
	c.patchJump(endJump)
}

func (c *Compiler) binary(canAssign bool) {
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

func (c *Compiler) str(canAssign bool) {
	c.emitConstant(ObjVal{Object: CreateStringObj(c.Ps.previous.Lexeme[1 : len(c.Ps.previous.Lexeme)-1])})
}

func (c *Compiler) grouping(canAssign bool) {
	c.expression()
	c.consume(TOKEN_RIGHT_PAREN, "Expected ')' after expression ")
}

func (c *Compiler) unary(canAssign bool) {
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

func (c *Compiler) variable(canAssign bool) {
	c.namedVariable(c.Ps.previous, canAssign)
}

func (c *Compiler) namedVariable(name Token, canAssign bool) {
	arg := c.resolveLocal(name)
	var getOp, setOp byte

	if arg != -1 {
		getOp = OP_GET_LOCAL
		setOp = OP_SET_LOCAL
	} else {
		arg = int(c.identifierConstant(name))
		getOp = OP_GET_GLOBAL
		setOp = OP_SET_GLOBAL
	}

	if canAssign && c.match(TOKEN_EQUAL) {
		c.expression()
		c.emitBytes(setOp, byte(arg))
	} else {
		c.emitBytes(getOp, byte(arg))
	}
}

func (c *Compiler) resolveLocal(name Token) int {
	for i := c.LocalCount - 1; i >= 0; i-- {
		local := c.Locals[i]
		if identifiersEqual(name, local.name) {
			if local.depth == -1 {
				c.error("Can't read local variable in its own initializer.")
			}
			return i
		}
	}
	return -1
}

func (c *Compiler) emitConstant(val Value) {
	c.emitBytes(OP_CONSTANT, c.makeConstant(val))
}

func (c *Compiler) makeConstant(val Value) byte {
	constant := c.Function.chunk.AddConstant(val)
	if constant > 255 {
		c.error("Too many constants in one chunk")
		return 0
	}
	return byte(constant)

}

func (c *Compiler) emitByte(b byte) {
	c.Function.chunk.Write(b, c.Ps.previous.Line)
}

func (c *Compiler) endCompiler() *ObjFunction {
	function := c.Function
	c.emitReturn()
	if DEBUG_PRINT_CODE {
		if !c.Ps.hadError {
			funcName := "<script>"
			if function.name != nil {
				funcName = function.name.Characters
			}
			DisassembleChunk(&c.Function.chunk, funcName)
		}
	}

	return function

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
		TOKEN_AND:           {nil, c.and_, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {c.literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {c.literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, c.or_, PREC_NONE},
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

func identifiersEqual(a Token, b Token) bool {
	return a.Lexeme == b.Lexeme
}
