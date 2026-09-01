package plan

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func TestCompileCallableAssignsStableLocalSlots(t *testing.T) {
	p := parser.NewParser(parser.NewLexer(`public func add(int $left, int $right): int { $sum = $left + $right return $sum }`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatal(errors)
	}
	method := program.Statements[0].(*parser.MethodStatement)
	compiled := CompileMethod(method, false)
	if compiled.RequiredCount != 2 || compiled.ParameterCount != 2 || len(compiled.Slots) != 3 {
		t.Fatalf("compiled plan = %#v", compiled)
	}
	if compiled.NameSlots["left"] != 0 || compiled.NameSlots["right"] != 1 || compiled.NameSlots["sum"] != 2 {
		t.Fatalf("slot map = %#v", compiled.NameSlots)
	}
	if len(compiled.IdentifierSlots) < 5 {
		t.Fatalf("identifier annotations = %#v", compiled.IdentifierSlots)
	}
}

func TestCompileCallableKeepsNestedClosureIndependent(t *testing.T) {
	p := parser.NewParser(parser.NewLexer(`public func outer(int $value) { $fn = func(int $inner) { return $inner + $value } return $fn }`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatal(errors)
	}
	compiled := CompileMethod(program.Statements[0].(*parser.MethodStatement), false)
	if _, exists := compiled.NameSlots["inner"]; exists {
		t.Fatalf("nested closure parameter leaked into outer plan: %#v", compiled.NameSlots)
	}
}
