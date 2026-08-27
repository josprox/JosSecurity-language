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
