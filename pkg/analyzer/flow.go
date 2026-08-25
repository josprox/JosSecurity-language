package analyzer

import "github.com/jossecurity/joss/pkg/parser"

// expressionTerminatesCallable is a side-effect-free control-flow query. It
// complements semantic inference without emitting duplicate diagnostics.
func expressionTerminatesCallable(expression parser.Expression) bool {
	switch node := expression.(type) {
	case *parser.BlockExpression:
		return blockTerminatesCallable(node.Block)
	case *parser.TernaryExpression:
		return node.True != nil && node.False != nil &&
			expressionTerminatesCallable(node.True) && expressionTerminatesCallable(node.False)
	case *parser.MatchExpression:
		if len(node.Arms) == 0 {
			return false
		}
		hasDefault := false
		for _, arm := range node.Arms {
			hasDefault = hasDefault || arm.IsDefault
			if !expressionTerminatesCallable(arm.Value) {
				return false
			}
		}
		return hasDefault
	default:
		return false
	}
}

func statementTerminatesCallable(statement parser.Statement) bool {
	switch node := statement.(type) {
	case *parser.ReturnStatement, *parser.ThrowStatement:
		return true
	case *parser.ExpressionStatement:
		return expressionTerminatesCallable(node.Expression)
	case *parser.TryCatchStatement:
		return node.TryBlock != nil && node.CatchBlock != nil &&
			blockTerminatesCallable(node.TryBlock) && blockTerminatesCallable(node.CatchBlock)
	default:
		return false
	}
}

func blockTerminatesCallable(block *parser.BlockStatement) bool {
	if block == nil {
		return false
	}
	for _, statement := range block.Statements {
		if statementTerminatesCallable(statement) {
			return true
		}
	}
	return false
}
