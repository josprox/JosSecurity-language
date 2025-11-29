package parser

import (
	"bytes"
	"strings"
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

type Identifier struct {
	Token Token // The token.VAR ($)
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

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

type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type CallExpression struct {
	Token     Token      // The '(' token
	Function  Expression // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
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

type TernaryExpression struct {
	Token     Token // ?
	Condition Expression
	True      *BlockStatement
	False     *BlockStatement
}

func (te *TernaryExpression) expressionNode()      {}
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TernaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(te.Condition.String())
	out.WriteString(") ? ")
	out.WriteString(te.True.String())
	out.WriteString(" : ")
	out.WriteString(te.False.String())
	return out.String()
}

type InfixExpression struct {
	Token    Token // Operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

type PrefixExpression struct {
	Token    Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

type Boolean struct {
	Token Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type ArrayLiteral struct {
	Token    Token // '['
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type MapLiteral struct {
	Token Token // '{'
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type IndexExpression struct {
	Token Token // '['
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
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

type FunctionLiteral struct {
	Token      Token // FUNCTION
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(fl.TokenLiteral() + "(")
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

type NewExpression struct {
	Token     Token // NEW
	Class     *Identifier
	Arguments []Expression
}

func (ne *NewExpression) expressionNode()      {}
func (ne *NewExpression) TokenLiteral() string { return ne.Token.Literal }
func (ne *NewExpression) String() string {
	var out bytes.Buffer
	out.WriteString("new ")
	out.WriteString(ne.Class.String())
	out.WriteString("(")
	args := []string{}
	for _, a := range ne.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

type MemberExpression struct {
	Token    Token // DOT
	Left     Expression
	Property *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	var out bytes.Buffer
	out.WriteString(me.Left.String())
	out.WriteString(".")
	out.WriteString(me.Property.String())
	return out.String()
}

type AssignExpression struct {
	Token Token // =
	Left  Expression
	Value Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ae.Left.String())
	out.WriteString(" = ")
	out.WriteString(ae.Value.String())
	return out.String()
}

type IssetExpression struct {
	Token     Token // ISSET
	Arguments []Expression
}

func (ie *IssetExpression) expressionNode()      {}
func (ie *IssetExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IssetExpression) String() string {
	var out bytes.Buffer
	out.WriteString("isset(")
	args := []string{}
	for _, a := range ie.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

type EmptyExpression struct {
	Token    Token // EMPTY
	Argument Expression
}

func (ee *EmptyExpression) expressionNode()      {}
func (ee *EmptyExpression) TokenLiteral() string { return ee.Token.Literal }
func (ee *EmptyExpression) String() string {
	var out bytes.Buffer
	out.WriteString("empty(")
	out.WriteString(ee.Argument.String())
	out.WriteString(")")
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

type BlockExpression struct {
	Token Token // {
	Block *BlockStatement
}

func (be *BlockExpression) expressionNode()      {}
func (be *BlockExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BlockExpression) String() string {
	return be.Block.String()
}
