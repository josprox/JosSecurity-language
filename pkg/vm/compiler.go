package vm

import (
	"fmt"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

type Compiler struct {
	chunk *Chunk
	slots map[string]int
}

func NewCompiler() *Compiler {
	return &Compiler{
		chunk: NewChunk(),
		slots: make(map[string]int),
	}
}

func (c *Compiler) getSlot(name string) int {
	clean := strings.TrimPrefix(name, "$")
	if slot, ok := c.slots[clean]; ok {
		return slot
	}
	slot := len(c.slots)
	c.slots[clean] = slot
	return slot
}

func (c *Compiler) Compile(program *parser.Program) (*Chunk, error) {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}
	c.chunk.WriteOp(OpHalt, 0)
	return c.chunk, nil
}

func (c *Compiler) compileStatement(stmt parser.Statement) error {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		slot := c.getSlot(s.Name.Value)
		if s.Value != nil {
			if err := c.compileExpression(s.Value); err != nil {
				return err
			}
		} else {
			c.chunk.WriteConst(IntVal(0), s.Token.Line)
		}
		c.chunk.WriteOp(OpStoreLocal, s.Token.Line)
		c.chunk.Write(byte(slot), s.Token.Line)

	case *parser.ExpressionStatement:
		if err := c.compileExpression(s.Expression); err != nil {
			return err
		}
		c.chunk.WriteOp(OpPop, 0)

	case *parser.WhileStatement:
		loopStart := len(c.chunk.Code)
		if err := c.compileExpression(s.Condition); err != nil {
			return err
		}
		jumpOut := len(c.chunk.Code)
		c.chunk.WriteOp(OpJumpIfFalse, 0)
		c.chunk.Write(0, 0)
		c.chunk.Write(0, 0)

		for _, bodyStmt := range s.Body.Statements {
			if err := c.compileStatement(bodyStmt); err != nil {
				return err
			}
		}

		loopOffset := loopStart - (len(c.chunk.Code) + 3)
		c.chunk.WriteOp(OpJump, 0)
		c.chunk.Write(byte(int16(loopOffset)>>8), 0)
		c.chunk.Write(byte(int16(loopOffset)&0xFF), 0)

		exitOffset := len(c.chunk.Code) - (jumpOut + 3)
		c.chunk.Code[jumpOut+1] = byte(int16(exitOffset) >> 8)
		c.chunk.Code[jumpOut+2] = byte(int16(exitOffset) & 0xFF)

	case *parser.ReturnStatement:
		if s.ReturnValue != nil {
			if err := c.compileExpression(s.ReturnValue); err != nil {
				return err
			}
		} else {
			c.chunk.WriteConst(NullVal(), s.Token.Line)
		}
		c.chunk.WriteOp(OpReturn, s.Token.Line)

	default:
		return fmt.Errorf("unsupported statement in prototype VM compiler: %T", stmt)
	}
	return nil
}

func (c *Compiler) compileExpression(expr parser.Expression) error {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		c.chunk.WriteConst(IntVal(e.Value), e.Token.Line)

	case *parser.FloatLiteral:
		c.chunk.WriteConst(FloatVal(e.Value), e.Token.Line)

	case *parser.Boolean:
		c.chunk.WriteConst(BoolVal(e.Value), e.Token.Line)

	case *parser.StringLiteral:
		c.chunk.WriteConst(StringVal(e.Value), e.Token.Line)

	case *parser.NullLiteral:
		c.chunk.WriteConst(NullVal(), e.Token.Line)

	case *parser.Identifier:
		if e.Value == "null" || e.Value == "nil" {
			c.chunk.WriteConst(NullVal(), e.Token.Line)
			return nil
		}
		if e.Value == "true" {
			c.chunk.WriteConst(BoolVal(true), e.Token.Line)
			return nil
		}
		if e.Value == "false" {
			c.chunk.WriteConst(BoolVal(false), e.Token.Line)
			return nil
		}
		slot := c.getSlot(e.Value)
		c.chunk.WriteOp(OpLoadLocal, e.Token.Line)
		c.chunk.Write(byte(slot), e.Token.Line)

	case *parser.AssignExpression:
		if ident, ok := e.Left.(*parser.Identifier); ok {
			slot := c.getSlot(ident.Value)
			if err := c.compileExpression(e.Value); err != nil {
				return err
			}
			c.chunk.WriteOp(OpStoreLocal, ident.Token.Line)
			c.chunk.Write(byte(slot), ident.Token.Line)
			c.chunk.WriteOp(OpLoadLocal, ident.Token.Line)
			c.chunk.Write(byte(slot), ident.Token.Line)
			return nil
		}
		return fmt.Errorf("unsupported assignment target in VM compiler: %T", e.Left)

	case *parser.InfixExpression:
		if err := c.compileExpression(e.Left); err != nil {
			return err
		}
		if err := c.compileExpression(e.Right); err != nil {
			return err
		}
		switch e.Operator {
		case "+":
			c.chunk.WriteOp(OpAddInt, e.Token.Line)
		case "-":
			c.chunk.WriteOp(OpSubInt, e.Token.Line)
		case "*":
			c.chunk.WriteOp(OpMulInt, e.Token.Line)
		case "/":
			c.chunk.WriteOp(OpDivFloat, e.Token.Line)
		case "%":
			c.chunk.WriteOp(OpModInt, e.Token.Line)
		case "<":
			c.chunk.WriteOp(OpCmpLt, e.Token.Line)
		case "<=":
			c.chunk.WriteOp(OpCmpLe, e.Token.Line)
		case ">":
			c.chunk.WriteOp(OpCmpGt, e.Token.Line)
		case ">=":
			c.chunk.WriteOp(OpCmpGe, e.Token.Line)
		case "==", "===":
			c.chunk.WriteOp(OpCmpEq, e.Token.Line)
		case "!=", "!==":
			c.chunk.WriteOp(OpCmpNe, e.Token.Line)
		default:
			return fmt.Errorf("unsupported operator in VM compiler: %s", e.Operator)
		}

	case *parser.PostfixExpression:
		if ident, ok := e.Left.(*parser.Identifier); ok && e.Operator == "++" {
			slot := c.getSlot(ident.Value)
			c.chunk.WriteOp(OpLoadLocal, e.Token.Line)
			c.chunk.Write(byte(slot), e.Token.Line)
			c.chunk.WriteConst(IntVal(1), e.Token.Line)
			c.chunk.WriteOp(OpAddInt, e.Token.Line)
			c.chunk.WriteOp(OpStoreLocal, e.Token.Line)
			c.chunk.Write(byte(slot), e.Token.Line)
			c.chunk.WriteOp(OpLoadLocal, e.Token.Line)
			c.chunk.Write(byte(slot), e.Token.Line)
			return nil
		}
		return fmt.Errorf("unsupported postfix expression in VM compiler: %T", e)

	case *parser.PrefixExpression:
		if e.Operator == "-" {
			if err := c.compileExpression(e.Right); err != nil {
				return err
			}
			c.chunk.WriteOp(OpNegInt, e.Token.Line)
			return nil
		}
		return fmt.Errorf("unsupported prefix operator in VM compiler: %s", e.Operator)

	default:
		return fmt.Errorf("unsupported expression in VM compiler: %T", expr)
	}
	return nil
}
