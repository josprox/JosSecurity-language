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
	ASSIGN:       ASSIGNMENT,
	QUESTION:     TERNARY,
	PLUS:         SUM,
	MINUS:        SUM,
	SLASH:        PRODUCT,
	ASTERISK:     PRODUCT,
	LT:           LESSGREATER,
	GT:           LESSGREATER,
	EQ:           EQUALS,
	NOT_EQ:       EQUALS,
	LTE:          LESSGREATER,
	GTE:          LESSGREATER,
	SHIFT_LEFT:   SHIFT,
	SHIFT_RIGHT:  SHIFT,
	LPAREN:       CALL,
	LBRACKET:     INDEX,
	DOT:          INDEX,
	ARROW:        INDEX,
	DOUBLE_COLON: INDEX,
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
	p.registerPrefix(FLOAT, p.parseFloatLiteral)
	p.registerPrefix(STRING, p.parseStringLiteral)
	p.registerPrefix(TRUE, p.parseBoolean)
	p.registerPrefix(FALSE, p.parseBoolean)
	p.registerPrefix(LPAREN, p.parseGroupedExpression)
	p.registerPrefix(LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(LBRACE, p.parseBraceExpression) // Maps { key: val } or Blocks { stmt; }
	p.registerPrefix(NEW, p.parseNewExpression)
	p.registerPrefix(NEW, p.parseNewExpression)
	p.registerPrefix(THIS, p.parseIdentifier)
	p.registerPrefix(ISSET, p.parseIssetExpression)
	p.registerPrefix(ISSET, p.parseIssetExpression)
	p.registerPrefix(EMPTY, p.parseEmptyExpression)
	p.registerPrefix(BANG, p.parsePrefixExpression)
	p.registerPrefix(MINUS, p.parsePrefixExpression)
	p.registerPrefix(FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[TokenType]infixParseFn)
	p.registerInfix(PLUS, p.parseInfixExpression)
	p.registerInfix(MINUS, p.parseInfixExpression)
	p.registerInfix(SLASH, p.parseInfixExpression)
	p.registerInfix(ASTERISK, p.parseInfixExpression)
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
	p.registerInfix(ARROW, p.parseMemberExpression)
	p.registerInfix(DOUBLE_COLON, p.parseMemberExpression)
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
	if p.curToken.Type == FUNCTION {
		return p.parseMethodStatement()
	}
	if p.curToken.Type == IMPORT {
		return p.parseImportStatement()
	}
	if p.curToken.Type == ECHO || p.curToken.Type == PRINT {
		return p.parseEchoStatement()
	}
	if p.curToken.Type == IF {
		return p.parseIfStatement()
	}
	if p.curToken.Type == WHILE {
		return p.parseWhileStatement()
	}
	if p.curToken.Type == DO {
		return p.parseDoWhileStatement()
	}
	if p.curToken.Type == TRY {
		return p.parseTryCatchStatement()
	}
	if p.curToken.Type == THROW {
		return p.parseThrowStatement()
	}
	if p.curToken.Type == RETURN {
		return p.parseReturnStatement()
	}
	// Check for variable declaration: type $name = value
	if p.curToken.Type == IDENT && p.peekToken.Type == VAR {
		return p.parseLetStatement()
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}

	p.nextToken()

	if p.curToken.Type == SEMICOLON {
		return stmt
	}

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseClassStatement() *ClassStatement {
	stmt := &ClassStatement{Token: p.curToken}

	if !p.expectPeek(IDENT) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekToken.Type == EXTENDS {
		p.nextToken()
		p.nextToken()
		stmt.SuperClass = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

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
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	// Parse parameters
	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(LBRACE) {
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

func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
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

func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
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

func (p *Parser) parseBraceExpression() Expression {
	// This could be a MapLiteral { key: val } or a BlockExpression { stmt; }
	// We need to look ahead.
	// But parsing an expression might consume tokens.
	// Strategy: Parse the first statement/expression.
	// If it's an expression and followed by COLON, it's a Map.
	// Otherwise, it's a Block.

	// However, p.parseExpression() consumes tokens.
	// If we parse an expression, we can't easily "unparse" it to treat it as a statement if it was a statement.
	// But in Joss, ExpressionStatement is a Statement.
	// So if we parse a Block, we expect Statements.
	// If we parse a Map, we expect Expressions.

	// Let's try a heuristic:
	// If the first token is RBRACE, it's empty map or empty block. Let's say empty Map {} (or empty block, doesn't matter much for logic, but empty map is more common as value).
	if p.peekToken.Type == RBRACE {
		p.nextToken()
		return &MapLiteral{Token: p.curToken, Pairs: make(map[Expression]Expression)}
	}

	// If we are in a block, we might see statements like 'return', 'if', 'for', 'var'.
	// These are NOT valid starts for a Map key (expression).
	// So if we see a keyword that starts a statement but not an expression, it MUST be a Block.
	if isStatementStart(p.peekToken.Type) {
		return p.parseBlockExpression()
	}

	// If it's an identifier or literal, it could be either.
	// { "a": 1 } -> Map
	// { "a" } -> Block (ExpressionStatement "a")
	// { $a = 1 } -> Block (AssignExpression is Expression, so ExpressionStatement)
	// { $a } -> Block

	// We have to parse the first "thing".
	// Since we are inside parseExpression (prefix position), we are expected to return an Expression.
	// If we return a BlockExpression, that's fine.

	// Let's rely on the fact that a Map key MUST be followed by a COLON.
	// But we can't parse the expression to check for colon without consuming it.
	// And if we consume it, we can't easily put it back into a BlockStatement.

	// ALTERNATIVE:
	// Use backtracking? No, expensive.
	// Use the fact that `parseBlockStatement` exists.
	// We can try to parse as a Block.
	// But `parseBlockStatement` expects statements.
	// `parseStatement` calls `parseExpressionStatement` which calls `parseExpression`.

	// Let's try this:
	// 1. Check if it looks like a Map.
	//    Map keys are usually literals or identifiers.
	//    If we see `IDENT :` or `STRING :` or `INT :`, it is definitely a Map.
	//    If we see `IDENT` then `ASSIGN` ($a = 1), it is definitely a Block (assignment).
	//    If we see `IDENT` then `SEMICOLON` or `NEWLINE`, it is a Block.

	// Peek 2 tokens ahead?
	// p.peekToken is next. p.l.PeekToken() ? We don't have double peek in this parser struct easily (only cur and peek).
	// We can add `peekTwice`?

	// Let's assume it's a BlockExpression by default, UNLESS we see `Key : Value`.
	// But we can't verify `Key : Value` without parsing Key.

	// Let's change the logic:
	// We parse a BlockStatement.
	// Inside the block parsing, if we find that the FIRST statement is an ExpressionStatement,
	// AND the next token is COLON, then we realize "Oh, this was actually a Map!".
	// Then we convert that ExpressionStatement's expression into the first Key of the map.
	// This is tricky but possible.

	block := &BlockStatement{Token: p.curToken, Statements: []Statement{}}
	p.nextToken() // consume LBRACE

	// Check for empty
	if p.curToken.Type == RBRACE {
		// Empty {} -> Map (standard convention in dynamic langs)
		return &MapLiteral{Token: p.curToken, Pairs: make(map[Expression]Expression)}
	}

	// Parse first statement
	// Handle NEWLINEs
	for p.curToken.Type == NEWLINE {
		p.nextToken()
	}
	if p.curToken.Type == RBRACE {
		return &MapLiteral{Token: p.curToken, Pairs: make(map[Expression]Expression)}
	}

	firstStmt := p.parseStatement()

	// If the first statement is NOT an ExpressionStatement, it's definitely a Block.
	// e.g. { return 1; } or { if ... }
	exprStmt, isExpr := firstStmt.(*ExpressionStatement)
	if !isExpr {
		// It's a block. Continue parsing statements.
		if firstStmt != nil {
			block.Statements = append(block.Statements, firstStmt)
		}
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
		return &BlockExpression{Token: block.Token, Block: block}
	}

	// It IS an expression statement. Check for COLON.
	// p.parseStatement() usually consumes the semicolon/newline.
	// But `parseExpressionStatement` consumes the expression, then optionally semicolon.
	// If it was a Map key, it wouldn't have a semicolon.
	// But `parseExpression` stops at LOWEST precedence.

	// If the next token (current token after parseStatement finished?) is COLON?
	// Wait, `parseStatement` advances tokens.
	// If `parseExpression` stopped at COLON, then `parseExpressionStatement` would see COLON as next token?
	// `parseExpression` loop: `for ... precedence < p.peekPrecedence()`.
	// COLON is not an infix operator in the map above (precedences).
	// So `parseExpression` stops before COLON.
	// Then `parseExpressionStatement` checks for SEMICOLON/NEWLINE.
	// If it sees COLON, it does nothing and returns.
	// So `p.peekToken` (or `p.curToken`?) should be COLON.

	// Let's verify `parseExpressionStatement`:
	// stmt.Expression = p.parseExpression(LOWEST)
	// if p.peekToken.Type == SEMICOLON || p.peekToken.Type == NEWLINE { p.nextToken() }

	// If input is `{ "a": 1 }`
	// LBRACE consumed.
	// parseStatement called.
	// parseExpressionStatement called.
	// parseExpression("a") returns StringLiteral.
	// peekToken is COLON.
	// parseExpressionStatement sees COLON != SEMI/NEWLINE. It does NOT consume it.
	// Returns ExpressionStatement.
	// Now inside `parseBraceExpression`:
	// p.peekToken should be COLON.

	if p.peekToken.Type == COLON {
		// It's a Map!
		// Convert firstStmt to Key.
		mapLit := &MapLiteral{Token: block.Token, Pairs: make(map[Expression]Expression)}
		key := exprStmt.Expression

		p.nextToken() // curToken is :
		p.nextToken() // curToken is start of value

		val := p.parseExpression(LOWEST)
		mapLit.Pairs[key] = val

		// Continue parsing map
		for !p.peekTokenIs(RBRACE) {
			if p.peekTokenIs(NEWLINE) {
				p.nextToken()
			}
			if p.peekTokenIs(COMMA) {
				p.nextToken()
			}
			if p.peekTokenIs(NEWLINE) {
				p.nextToken()
			}
			if p.peekTokenIs(RBRACE) {
				break
			}

			p.nextToken() // start of next key
			key := p.parseExpression(LOWEST)

			if p.peekTokenIs(NEWLINE) {
				p.nextToken()
			}

			if !p.expectPeek(COLON) {
				return nil
			}
			p.nextToken()
			val := p.parseExpression(LOWEST)
			mapLit.Pairs[key] = val
		}

		if !p.expectPeek(RBRACE) {
			return nil
		}
		return mapLit
	}

	// Not a map. It's a Block.
	if firstStmt != nil {
		block.Statements = append(block.Statements, firstStmt)
	}
	p.nextToken() // Move past the last token of first statement

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

	return &BlockExpression{Token: block.Token, Block: block}
}

func (p *Parser) parseBlockExpression() *BlockExpression {
	block := p.parseBlockStatement()
	return &BlockExpression{Token: block.Token, Block: block}
}

func isStatementStart(t TokenType) bool {
	switch t {
	case RETURN, IF, VAR, FOREACH, WHILE, DO, TRY, THROW, ECHO, PRINT:
		return true
	}
	return false
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

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
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

	for p.peekToken.Type == COMMA || p.peekToken.Type == NEWLINE {
		if p.peekToken.Type == NEWLINE {
			p.nextToken()
			// Check if we hit RPAREN after newline
			if p.peekToken.Type == RPAREN {
				break
			}
			// If no comma after newline, we assume comma insertion or just continue if next is expression
			if p.peekToken.Type != COMMA {
				// Optional: check if next token is start of expression?
				// For now, let's assume if it's not comma, it might be next arg (if comma is optional)
				// But standard Joss requires comma.
				// However, let's be safe and check for comma.
				if p.peekToken.Type != COMMA {
					// If not comma, maybe we should continue loop to let parseExpression handle it?
					// Or break?
					// Let's just continue and let the loop condition handle it (it won't match COMMA).
					// But we are inside the loop.
				}
			}
		}

		if p.peekToken.Type == COMMA {
			p.nextToken()
			// Allow newline after comma
			for p.peekToken.Type == NEWLINE {
				p.nextToken()
			}
			p.nextToken() // Advance to start of expression
			args = append(args, p.parseExpression(LOWEST))
		}
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

func (p *Parser) parseWhileStatement() *WhileStatement {
	stmt := &WhileStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseDoWhileStatement() *DoWhileStatement {
	stmt := &DoWhileStatement{Token: p.curToken}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if !p.expectPeek(WHILE) {
		return nil
	}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTryCatchStatement() *TryCatchStatement {
	stmt := &TryCatchStatement{Token: p.curToken}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.TryBlock = p.parseBlockStatement()

	if !p.expectPeek(CATCH) {
		return nil
	}
	stmt.CatchToken = p.curToken

	if !p.expectPeek(LPAREN) {
		return nil
	}

	// Expect variable: $e
	if !p.expectPeek(VAR) {
		return nil
	}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.CatchVar = p.curToken.Literal

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.CatchBlock = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseThrowStatement() *ThrowStatement {
	stmt := &ThrowStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}

	return stmt
}
func (p *Parser) parseIfStatement() *IfStatement {
	stmt := &IfStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == ELSE {
		p.nextToken()

		if p.peekToken.Type == IF {
			// else if... treat as nested if in Alternative?
			// Or just parseIfStatement again?
			// Standard way: else { if ... } or specific ElseIf support.
			// Let's support simple else block first.
			// If we want else if, we can check if next token is IF.

			// Actually, if peek is IF, we can just parse it as the Alternative (which is a Statement).
			// But parseBlockStatement expects LBRACE.
			// So "else if" is usually parsed recursively or as a block containing an if.

			// Let's handle simple ELSE { BLOCK } first.
			if !p.expectPeek(LBRACE) {
				// Support "else if" by checking if next is IF?
				// If p.peekToken.Type == IF { ... }
				// For now, let's stick to strict ELSE { }
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		} else {
			if !p.expectPeek(LBRACE) {
				return nil
			}
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}
