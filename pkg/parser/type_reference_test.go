package parser

import "testing"

func TestParserNormalizesUnionAndNullableTypes(t *testing.T) {
	p := NewParser(NewLexer(`public func choose(int|string $value): string? { return null }
int? $count = null`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	function := program.Statements[0].(*MethodStatement)
	if function.Parameters[0].Type.Literal != "int|string" || function.ReturnType.Literal != "string|null" {
		t.Fatalf("function types = %q -> %q", function.Parameters[0].Type.Literal, function.ReturnType.Literal)
	}
	declaration := program.Statements[1].(*LetStatement)
	if declaration.Token.Literal != "int|null" {
		t.Fatalf("declaration type = %q", declaration.Token.Literal)
	}
}

func TestParserParsesDecimalLiteralsAndDeclarations(t *testing.T) {
	p := NewParser(NewLexer(`decimal $precio = 0.50m
$total = 100.25M`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	letStmt := program.Statements[0].(*LetStatement)
	if letStmt.Token.Literal != "decimal" {
		t.Fatalf("expected decimal type, got %q", letStmt.Token.Literal)
	}
	decLit, ok := letStmt.Value.(*DecimalLiteral)
	if !ok || decLit.Value.String() != "0.5" {
		t.Fatalf("expected 0.5 decimal literal, got %v", letStmt.Value)
	}

	assignStmt := program.Statements[1].(*ExpressionStatement)
	assignExpr := assignStmt.Expression.(*AssignExpression)
	decLit2, ok := assignExpr.Value.(*DecimalLiteral)
	if !ok || decLit2.Value.String() != "100.25" {
		t.Fatalf("expected 100.25 decimal literal, got %v", assignExpr.Value)
	}
}
