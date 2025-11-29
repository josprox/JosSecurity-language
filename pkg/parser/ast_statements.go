package parser

import (
	"bytes"
	"strings"
)

// LetStatement: string $x = "foo"
type LetStatement struct {
	Token Token // The token.IDENT (e.g. string, int)
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.Token.Literal + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ExpressionStatement struct {
	Token      Token // The first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type ClassStatement struct {
	Token      Token // CLASS
	Name       *Identifier
	SuperClass *Identifier
	Body       *BlockStatement
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString("class ")
	out.WriteString(cs.Name.String())
	if cs.SuperClass != nil {
		out.WriteString(" extends ")
		out.WriteString(cs.SuperClass.String())
	}
	out.WriteString(" ")
	out.WriteString(cs.Body.String())
	return out.String()
}

type BlockStatement struct {
	Token      Token // {
	Statements []Statement
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

type EchoStatement struct {
	Token Token // 'echo' or 'print'
	Value Expression
}

func (es *EchoStatement) statementNode()       {}
func (es *EchoStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EchoStatement) String() string {
	var out bytes.Buffer
	out.WriteString(es.Token.Literal + " ")
	if es.Value != nil {
		out.WriteString(es.Value.String())
	}
	return out.String()
}

type InitStatement struct {
	Token      Token       // INIT
	Name       *Identifier // main
	Parameters []*Identifier
	Body       *BlockStatement
}

func (is *InitStatement) statementNode()       {}
func (is *InitStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InitStatement) String() string {
	var out bytes.Buffer
	out.WriteString("Init ")
	out.WriteString(is.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range is.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(is.Body.String())
	return out.String()
}

type ForeachStatement struct {
	Token    Token // 'foreach'
	Iterable Expression
	Value    string // The variable name, e.g. "val" in "as $val"
	Body     *BlockStatement
}

func (fs *ForeachStatement) statementNode()       {}
func (fs *ForeachStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForeachStatement) String() string {
	var out bytes.Buffer
	out.WriteString("foreach (")
	out.WriteString(fs.Iterable.String())
	out.WriteString(" as $")
	out.WriteString(fs.Value)
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

type ImportStatement struct {
	Token Token // IMPORT
	Path  string
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString("Import \"")
	out.WriteString(is.Path)
	out.WriteString("\"")
	return out.String()
}

type MethodStatement struct {
	Token      Token // FUNCTION
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (ms *MethodStatement) statementNode()       {}
func (ms *MethodStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *MethodStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ms.TokenLiteral() + " ")
	out.WriteString(ms.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range ms.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(ms.Body.String())
	return out.String()
}

type WhileStatement struct {
	Token     Token // WHILE
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while (")
	out.WriteString(ws.Condition.String())
	out.WriteString(") ")
	out.WriteString(ws.Body.String())
	return out.String()
}

type DoWhileStatement struct {
	Token     Token // DO
	Condition Expression
	Body      *BlockStatement
}

func (dws *DoWhileStatement) statementNode()       {}
func (dws *DoWhileStatement) TokenLiteral() string { return dws.Token.Literal }
func (dws *DoWhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("do ")
	out.WriteString(dws.Body.String())
	out.WriteString(" while (")
	out.WriteString(dws.Condition.String())
	out.WriteString(");")
	return out.String()
}

type TryCatchStatement struct {
	Token      Token // TRY
	TryBlock   *BlockStatement
	CatchToken Token  // CATCH
	CatchVar   string // The variable name for the error, e.g. "e"
	CatchBlock *BlockStatement
}

func (tcs *TryCatchStatement) statementNode()       {}
func (tcs *TryCatchStatement) TokenLiteral() string { return tcs.Token.Literal }
func (tcs *TryCatchStatement) String() string {
	var out bytes.Buffer
	out.WriteString("try ")
	out.WriteString(tcs.TryBlock.String())
	out.WriteString(" catch ($")
	out.WriteString(tcs.CatchVar)
	out.WriteString(") ")
	out.WriteString(tcs.CatchBlock.String())
	return out.String()
}

type ThrowStatement struct {
	Token Token // THROW
	Value Expression
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	var out bytes.Buffer
	out.WriteString("throw ")
	if ts.Value != nil {
		out.WriteString(ts.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ReturnStatement struct {
	Token       Token // 'return'
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type IfStatement struct {
	Token       Token // 'if'
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}
