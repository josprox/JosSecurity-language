package vm

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func BenchmarkVMWhileLoopLarge(b *testing.B) {
	source := `
int $i = 0
while ($i < 10000) {
    $i++
}
`
	p := parser.NewParser(parser.NewLexer(source))
	prog := p.ParseProgram()
	comp := NewCompiler()
	chunk, err := comp.Compile(prog)
	if err != nil {
		b.Fatal(err)
	}

	vm := NewVM()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = vm.Run(chunk)
	}
}

func BenchmarkVMArithmetic(b *testing.B) {
	source := `
int $a = 10
int $b = 20
int $c = $a + $b * 2 - 5
return $c
`
	p := parser.NewParser(parser.NewLexer(source))
	prog := p.ParseProgram()
	comp := NewCompiler()
	chunk, err := comp.Compile(prog)
	if err != nil {
		b.Fatal(err)
	}

	vm := NewVM()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = vm.Run(chunk)
	}
}
