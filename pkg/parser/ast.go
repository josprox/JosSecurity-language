package parser

import (
	"bytes"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type Parameter struct {
	Type Token // Optional: string, int, etc.
	Name *Identifier
}

func (p *Parameter) String() string {
	res := ""
	if p.Type.Literal != "" && p.Type.Type != VAR { // VAR means $ which is not a type itself here
		res += p.Type.Literal + " "
	}
	res += p.Name.String()
	return res
}
