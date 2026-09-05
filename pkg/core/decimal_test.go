package core

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
	"github.com/shopspring/decimal"
)

func evalDecimalSource(source string) *Runtime {
	l := parser.NewLexer(source)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	rt := NewRuntime()
	rt.Execute(program)
	return rt
}

func TestDecimalExactArithmetic(t *testing.T) {
	source := `
decimal $a = 0.10m
decimal $b = 0.20m
decimal $c = $a + $b
$exactMatch = ($c == 0.30m)
`
	p := evalDecimalSource(source)
	match, ok := p.Variables["exactMatch"].(bool)
	if !ok || !match {
		t.Fatalf("expected exact match 0.10m + 0.20m == 0.30m, got %v", p.Variables["c"])
	}
}

func TestDecimalBankingCalculation(t *testing.T) {
	source := `
decimal $saldo = 0.60m
decimal $precio = 0.10m + 0.20m
$puedeComprar = ($saldo >= $precio)
decimal $resto = $saldo - $precio
$esTreinta = ($resto == 0.30m)
`
	p := evalDecimalSource(source)
	puedeComprar, ok := p.Variables["puedeComprar"].(bool)
	if !ok || !puedeComprar {
		t.Fatalf("expected puedeComprar == true")
	}
	esTreinta, ok := p.Variables["esTreinta"].(bool)
	if !ok || !esTreinta {
		t.Fatalf("expected $resto == 0.30m, got %v", p.Variables["resto"])
	}
}

func TestDecimalBuiltinsAndCoercion(t *testing.T) {
	source := `
$val = decimal("45.67")
$isDec = is_decimal($val)
$isNum = is_numeric($val)
decimal $fromInt = 100
$exact = ($fromInt == 100.0m)
$str = "Total: " . $val
`
	p := evalDecimalSource(source)
	if isDec, ok := p.Variables["isDec"].(bool); !ok || !isDec {
		t.Fatalf("expected is_decimal to be true")
	}
	if isNum, ok := p.Variables["isNum"].(bool); !ok || !isNum {
		t.Fatalf("expected is_numeric to be true")
	}
	if exact, ok := p.Variables["exact"].(bool); !ok || !exact {
		t.Fatalf("expected decimal from int coercion equality")
	}
	if str, ok := p.Variables["str"].(string); !ok || str != "Total: 45.67" {
		t.Fatalf("expected string concatenation 'Total: 45.67', got %q", str)
	}
}

func TestDecimalOperationsAndComparisons(t *testing.T) {
	source := `
decimal $x = 10.50m
decimal $y = 2.0m
$mul = $x * $y
$div = $x / $y
$neg = -$x
$cmpSpaceship = ($x <=> $y)
$cmpLess = ($y < $x)
$cmpGreater = ($x > $y)
`
	p := evalDecimalSource(source)
	mulVal, ok := p.Variables["mul"].(decimal.Decimal)
	if !ok || !mulVal.Equal(decimal.NewFromFloat(21.0)) {
		t.Fatalf("expected 21.0, got %v", p.Variables["mul"])
	}
	divVal, ok := p.Variables["div"].(decimal.Decimal)
	if !ok || !divVal.Equal(decimal.NewFromFloat(5.25)) {
		t.Fatalf("expected 5.25, got %v", p.Variables["div"])
	}
	negVal, ok := p.Variables["neg"].(decimal.Decimal)
	if !ok || !negVal.Equal(decimal.NewFromFloat(-10.50)) {
		t.Fatalf("expected -10.50, got %v", p.Variables["neg"])
	}
	cmp, ok := p.Variables["cmpSpaceship"].(int64)
	if !ok || cmp != 1 {
		t.Fatalf("expected spaceship 1, got %v", p.Variables["cmpSpaceship"])
	}
	if less, ok := p.Variables["cmpLess"].(bool); !ok || !less {
		t.Fatalf("expected cmpLess true")
	}
	if greater, ok := p.Variables["cmpGreater"].(bool); !ok || !greater {
		t.Fatalf("expected cmpGreater true")
	}
}
