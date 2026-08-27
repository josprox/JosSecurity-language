package parser

import "testing"

func TestReferenceParameterAndCallArgument(t *testing.T) {
	lexer := NewLexer(`public func increment(ref int $value): int { $value = $value + 1 return $value }
$age = 20
increment(ref $age)`)
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	if issues := parser.Diagnostics(); len(issues) != 0 {
		t.Fatalf("parser diagnostics: %+v", issues)
	}
	method, ok := program.Statements[0].(*MethodStatement)
	if !ok || len(method.Parameters) != 1 || !method.Parameters[0].ByReference {
		t.Fatalf("expected one ref parameter, got %#v", program.Statements[0])
	}
	callStatement, ok := program.Statements[2].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected call statement, got %T", program.Statements[2])
	}
	call, ok := callStatement.Expression.(*CallExpression)
	if !ok || len(call.Arguments) != 1 {
		t.Fatalf("expected one call argument, got %#v", callStatement.Expression)
	}
	if _, ok := call.Arguments[0].(*ReferenceExpression); !ok {
		t.Fatalf("expected ref call argument, got %T", call.Arguments[0])
	}
}
