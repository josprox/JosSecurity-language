package vm

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func parseAndCompile(t *testing.T, source string) *Chunk {
	t.Helper()
	p := parser.NewParser(parser.NewLexer(source))
	prog := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	comp := NewCompiler()
	chunk, err := comp.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return chunk
}

func TestVMBasicArithmetic(t *testing.T) {
	chunk := parseAndCompile(t, `
int $a = 10
int $b = 20
int $c = $a + $b * 2
return $c
`)
	vm := NewVM()
	res, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("vm run error: %v", err)
	}
	if res.Kind != ValInt || res.Integer != 50 {
		t.Fatalf("expected int 50, got %v", res)
	}
}

func TestVMWhileLoop(t *testing.T) {
	chunk := parseAndCompile(t, `
int $i = 0
int $sum = 0
while ($i < 10) {
    $sum = $sum + $i
    $i++
}
return $sum
`)
	vm := NewVM()
	res, err := vm.Run(chunk)
	if err != nil {
		t.Fatalf("vm run error: %v", err)
	}
	if res.Kind != ValInt || res.Integer != 45 {
		t.Fatalf("expected sum 45, got %v", res)
	}
}
