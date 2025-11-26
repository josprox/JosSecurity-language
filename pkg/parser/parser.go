package parser

import (
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	ASSIGNMENT  // =
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	SHIFT       // << or >>
	PRODUCT     // *
	PREFIX      // -X or !X
	TERNARY     // ? :
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[TokenType]int{
	ASSIGN:      ASSIGNMENT,
	QUESTION:    TERNARY,
	PLUS:        SUM,
	LT:          LESSGREATER,
	GT:          LESSGREATER,
	EQ:          EQUALS,
	NOT_EQ:      EQUALS,
	LTE:         LESSGREATER,
	GTE:         LESSGREATER,
	SHIFT_LEFT:  SHIFT,
	SHIFT_RIGHT: SHIFT,
	LPAREN:      CALL,
	LBRACKET:    INDEX,
	DOT:         INDEX,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[TokenType]prefixParseFn)
	p.registerPrefix(IDENT, p.parseIdentifier)
	p.registerPrefix(VAR, p.parseVarExpression) // Handle $name
	p.registerPrefix(INT, p.parseIntegerLiteral)
	p.registerPrefix(STRING, p.parseStringLiteral)
	p.registerPrefix(TRUE, p.parseBoolean)
	p.registerPrefix(FALSE, p.parseBoolean)
	p.registerPrefix(LPAREN, p.parseGroupedExpression)
	p.registerPrefix(LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(NEW, p.parseNewExpression)
	p.registerPrefix(NEW, p.parseNewExpression)
	p.registerPrefix(THIS, p.parseIdentifier)
	p.registerPrefix(ISSET, p.parseIssetExpression)
	p.registerPrefix(EMPTY, p.parseEmptyExpression)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(LT, p.parseInfixExpression)
	p.registerInfix(GT, p.parseInfixExpression)
	p.registerInfix(EQ, p.parseInfixExpression)
	p.registerInfix(NOT_EQ, p.parseInfixExpression)
	p.registerInfix(LTE, p.parseInfixExpression)
	p.registerInfix(GTE, p.parseInfixExpression)
	p.registerInfix(SHIFT_LEFT, p.parseInfixExpression)
	p.registerInfix(SHIFT_RIGHT, p.parseInfixExpression)
	p.registerInfix(LPAREN, p.parseCallExpression)
	p.registerInfix(QUESTION, p.parseTernaryExpression)
	p.registerInfix(LBRACKET, p.parseIndexExpression)
	p.registerInfix(DOT, p.parseMemberExpression)
	p.registerInfix(ASSIGN, p.parseAssignExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for p.curToken.Type != EOF {
		if p.curToken.Type == NEWLINE {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() Statement {
	if p.curToken.Type == CLASS {
		return p.parseClassStatement()
	}
	if p.curToken.Type == INIT {
		return p.parseInitStatement()
	}
	if p.curToken.Type == FOREACH {
		return p.parseForeachStatement()
	}
	if p.curToken.Type == IMPORT {
		return p.parseImportStatement()
	}
	if p.curToken.Type == ECHO || p.curToken.Type == PRINT {
		return p.parseEchoStatement()
	}
	if p.curToken.Type == ECHO || p.curToken.Type == PRINT {
		return p.parseEchoStatement()
	}
	// Check for variable declaration: type $name = value
	if p.curToken.Type == IDENT && p.peekToken.Type == VAR {
		return p.parseLetStatement()
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	if !p.expectPeek(IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseClassBody()

	return stmt
}

func (p *Parser) parseClassBody() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	p.nextToken()

	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		if p.curToken.Type == NEWLINE {
			p.nextToken()
			continue
		}

		var stmt Statement
		if p.curToken.Type == FUNCTION {
			stmt = p.parseMethodStatement()
		} else if p.curToken.Type == INIT {
			stmt = p.parseInitStatement()
		} else if p.curToken.Type == IDENT && p.peekToken.Type == VAR { // Property: string $x
			stmt = p.parseLetStatement()
		} else {
			// Skip or error? For now skip to avoid infinite loop if unknown
			p.nextToken()
			continue
		}

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseInitStatement() *InitStatement {
	stmt := &InitStatement{Token: p.curToken}

	if !p.expectPeek(IDENT) { // main
		return nil
	}
	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	p.nextToken()

	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		if p.curToken.Type == NEWLINE {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseLetStatement() *LetStatement {
	stmt := &LetStatement{Token: p.curToken} // Type (string, int)

	if !p.expectPeek(VAR) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type == ASSIGN {
		p.nextToken()
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekToken.Type == SEMICOLON || p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == SEMICOLON || p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(SEMICOLON) && !p.peekTokenIs(NEWLINE) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseVarExpression() Expression {
	// Current token is VAR ($)
	// We expect next to be IDENT or THIS
	if p.peekToken.Type == THIS {
		p.nextToken()
		return &Identifier{Token: p.curToken, Value: "this"}
	}
	if !p.expectPeek(IDENT) {
		return nil
	}
	// Now curToken is IDENT
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() Expression {
	return &Boolean{Token: p.curToken, Value: p.curToken.Type == TRUE}
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseArrayLiteral() Expression {
	array := &ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseForeachStatement() *ForeachStatement {
	stmt := &ForeachStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(AS) {
		return nil
	}

	// Expect variable: $val
	// In parser, VAR is '$', then IDENT 'val'
	if !p.expectPeek(VAR) {
		return nil
	}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Value = p.curToken.Literal

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseImportStatement() *ImportStatement {
	stmt := &ImportStatement{Token: p.curToken}

	if !p.expectPeek(STRING) {
		return nil
	}

	stmt.Path = p.curToken.Literal

	if p.peekToken.Type == SEMICOLON || p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseEchoStatement() *EchoStatement {
	stmt := &EchoStatement{Token: p.curToken}

	p.nextToken() // consume ECHO/PRINT

	// Optional parentheses: echo("foo")
	if p.curToken.Type == LPAREN {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
		if p.peekToken.Type == RPAREN {
			p.nextToken()
		}
	} else {
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekToken.Type == SEMICOLON || p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseTernaryExpression(condition Expression) Expression {
	expression := &TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	// Check if it's a block ternary: ? {
	if p.peekToken.Type == LBRACE {
		p.nextToken()
		expression.True = p.parseBlockStatement()
	} else {
		// Value ternary: ? "A"
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		expression.True = &BlockStatement{
			Statements: []Statement{&ExpressionStatement{Expression: exp}},
		}
	}

	// Allow newlines before colon
	for p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	if !p.expectPeek(COLON) {
		return nil
	}

	// Allow newlines after colon
	for p.peekToken.Type == NEWLINE {
		p.nextToken()
	}

	if p.peekToken.Type == LBRACE {
		p.nextToken()
		expression.False = p.parseBlockStatement()
	} else {
		p.nextToken()
		exp := p.parseExpression(LOWEST)
		expression.False = &BlockStatement{
			Statements: []Statement{&ExpressionStatement{Expression: exp}},
		}
	}

	return expression
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}

	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(RPAREN) {
		return nil
	}

	return args
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

func (p *Parser) noPrefixParseFnError(t TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) parseMethodStatement() *MethodStatement {
	stmt := &MethodStatement{Token: p.curToken}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	// Expect variable: $param
	if p.curToken.Type == VAR {
		if !p.expectPeek(IDENT) {
			return nil
		}
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		if p.curToken.Type == VAR {
			if !p.expectPeek(IDENT) {
				return nil
			}
			ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
			identifiers = append(identifiers, ident)
		}
	}

	if !p.expectPeek(RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseNewExpression() Expression {
	exp := &NewExpression{Token: p.curToken}

	if !p.expectPeek(IDENT) {
		return nil
	}
	exp.Class = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	exp.Arguments = p.parseCallArguments()

	return exp
}

func (p *Parser) parseMemberExpression(left Expression) Expression {
	exp := &MemberExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(IDENT) {
		return nil
	}
	exp.Property = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

func (p *Parser) parseAssignExpression(left Expression) Expression {
	exp := &AssignExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Value = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseIssetExpression() Expression {
	exp := &IssetExpression{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	exp.Arguments = p.parseCallArguments()

	return exp
}

func (p *Parser) parseEmptyExpression() Expression {
	exp := &EmptyExpression{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	exp.Argument = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
