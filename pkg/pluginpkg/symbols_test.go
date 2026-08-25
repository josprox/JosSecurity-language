package pluginpkg

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func TestSymbolIndexPublishesCallableReturnTypes(t *testing.T) {
	p := parser.NewParser(parser.NewLexer(`func lookup(int $id): string|null { return null }
class Store { func count(): int { return 0 } }`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	index := BuildSymbolIndex(program, "sample", "1.0.0")
	if len(index.Functions) != 1 || index.Functions[0].ReturnType != "string|null" {
		t.Fatalf("function symbols = %#v", index.Functions)
	}
	if len(index.Classes) != 1 || len(index.Classes[0].Methods) != 1 || index.Classes[0].Methods[0].ReturnType != "int" {
		t.Fatalf("class symbols = %#v", index.Classes)
	}
}
