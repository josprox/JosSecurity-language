package core

import (
	"strings"
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func executeJossScript(script string) (*Runtime, error) {
	p := parser.NewParser(parser.NewLexer(script))
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return nil, &ScriptParseError{Errors: p.Errors()}
	}

	rt := NewRuntime()
	rt.Execute(prog)
	return rt, nil
}

type ScriptParseError struct {
	Errors []string
}

func (e *ScriptParseError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func TestCoreMethodDefaultParameters(t *testing.T) {
	src := `
	public class OrderCalculator {
		public func calculate(int $amount, float $tax = 0.16, int $discount = 0) {
			return $amount + ($amount * $tax) - $discount
		}
	}

	$calc = new OrderCalculator()
	$res1 = $calc->calculate(100)
	$res2 = $calc->calculate(100, 0.08, 10)
	`
	rt, err := executeJossScript(src)
	if err != nil {
		t.Fatalf("Script error: %v", err)
	}

	// 100 + (100 * 0.16) - 0 = 116
	res1 := rt.Variables["res1"]
	if res1 != float64(116) && res1 != int64(116) {
		t.Errorf("Expected res1 to be 116, got %v (%T)", res1, res1)
	}

	// 100 + (100 * 0.08) - 10 = 98
	res2 := rt.Variables["res2"]
	if res2 != float64(98) && res2 != int64(98) {
		t.Errorf("Expected res2 to be 98, got %v (%T)", res2, res2)
	}
}

func TestCoreTryCatchAndThrow(t *testing.T) {
	src := `
	$caughtMsg = ""
	$status = "pending"

	try {
		$status = "running"
		throw "Error de conexion con pasarela"
		$status = "finished"
	} catch ($err) {
		$caughtMsg = $err
		$status = "error_handled"
	}
	`
	rt, err := executeJossScript(src)
	if err != nil {
		t.Fatalf("Script error: %v", err)
	}

	if rt.Variables["status"] != "error_handled" {
		t.Errorf("Expected status to be 'error_handled', got %v", rt.Variables["status"])
	}

	if rt.Variables["caughtMsg"] != "Error de conexion con pasarela" {
		t.Errorf("Expected caughtMsg to match thrown error, got %v", rt.Variables["caughtMsg"])
	}
}

func TestCoreClassInheritance(t *testing.T) {
	src := `
	public class Animal {
		protected string $species = "unknown"

		public func speak() {
			return "Sonido generico de " . $this->species
		}
	}

	public class Dog extends Animal {
		public func constructor() {
			$this->species = "canino"
		}

		public func bark() {
			return "Guau!"
		}
	}

	$dog = new Dog()
	$msg1 = $dog->speak()
	$msg2 = $dog->bark()
	`
	rt, err := executeJossScript(src)
	if err != nil {
		t.Fatalf("Script error: %v", err)
	}

	if rt.Variables["msg1"] != "Sonido generico de canino" {
		t.Errorf("Expected inherited speak() to return 'Sonido generico de canino', got %v", rt.Variables["msg1"])
	}

	if rt.Variables["msg2"] != "Guau!" {
		t.Errorf("Expected bark() to return 'Guau!', got %v", rt.Variables["msg2"])
	}
}
