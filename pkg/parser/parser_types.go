package parser

import "strings"

func isTypeStart(token Token) bool {
	return token.Type == IDENT || token.Type == NULL || token.Type == NIL
}

func isTypeContinuation(tokenType TokenType) bool {
	return tokenType == TYPE_UNION || tokenType == QUESTION
}

// parseTypeReference consumes a source type from curToken. Union types use
// A|B and nullable shorthand uses T?. The AST stores the normalized spelling
// T|null, so every later phase consumes one representation.
func (p *Parser) parseTypeReference() Token {
	token := p.curToken
	parts := []string{token.Literal}
	for p.peekToken.Type == TYPE_UNION {
		p.nextToken()
		if !isTypeStart(p.peekToken) {
			p.addError(p.peekToken, "Se esperaba un tipo después de `|`.")
			break
		}
		p.nextToken()
		parts = append(parts, p.curToken.Literal)
	}
	if p.peekToken.Type == QUESTION {
		p.nextToken()
		parts = append(parts, "null")
	}
	token.Literal = strings.Join(parts, "|")
	return token
}
