package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func executeTypingSource(t *testing.T, source string) *Runtime {
	t.Helper()
	p := parser.NewParser(parser.NewLexer(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	runtime := NewRuntime()
	runtime.Execute(program)
	return runtime
}

func TestRuntimeInfersFirstAssignmentType(t *testing.T) {
	runtime := executeTypingSource(t, `$age = 20
$age = 30`)
	if runtime.VarTypes["age"] != "int" {
		t.Fatalf("age type = %q, want int", runtime.VarTypes["age"])
	}
}

func TestRuntimeRejectsInferredTypeChange(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil || !strings.Contains(fmt.Sprint(recovered), "Error de Tipado") {
			t.Fatalf("expected type panic, got %v", recovered)
		}
	}()
	executeTypingSource(t, `$age = 20
$age = "twenty"`)
}

func TestRuntimeAllowsExplicitMixedTypeChange(t *testing.T) {
	runtime := executeTypingSource(t, `let $value = 20
$value = "twenty"`)
	if runtime.Variables["value"] != "twenty" || runtime.VarTypes["value"] != "mixed" {
		t.Fatalf("unexpected dynamic value/type: %v / %q", runtime.Variables["value"], runtime.VarTypes["value"])
	}
}

func TestRuntimeKeepsTypedParameterTypeInsideFunction(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil || !strings.Contains(fmt.Sprint(recovered), "Error de Tipado") {
			t.Fatalf("expected typed parameter assignment panic, got %v", recovered)
		}
	}()
	executeTypingSource(t, `func change(int $value) { $value = "wrong" }
change(1)`)
}

func TestRuntimeRejectsWrongFunctionArity(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil || !strings.Contains(fmt.Sprint(recovered), "Arity Error") {
			t.Fatalf("expected arity panic, got %v", recovered)
		}
	}()
	executeTypingSource(t, `func add(int $a, int $b) { return $a + $b }
add(1)`)
}

func TestRuntimeRestoresParameterStateWhenBindingFails(t *testing.T) {
	runtime := executeTypingSource(t, `func accept(int $value, string $label) { return $value }`)
	defer runtime.Free()
	runtime.Variables["value"] = int64(7)
	runtime.VarTypes["value"] = "int"
	runtime.Variables["label"] = "before"
	runtime.VarTypes["label"] = "string"

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected argument type failure")
			}
		}()
		runtime.CallMethodEvaluated(runtime.Functions["accept"], nil, []interface{}{"wrong", "after"})
	}()

	if runtime.Variables["value"] != int64(7) || runtime.VarTypes["value"] != "int" ||
		runtime.Variables["label"] != "before" || runtime.VarTypes["label"] != "string" {
		t.Fatalf("failed binding leaked parameter state: %#v / %#v", runtime.Variables, runtime.VarTypes)
	}
}

func TestRuntimeSupportsDirectRecursion(t *testing.T) {
	runtime := executeTypingSource(t, `func factorial(int $n): int {
    ($n <= 1) ? { return 1 } : {}
    return $n * factorial($n - 1)
}
$result = factorial(5)`)
	defer runtime.Free()
	if runtime.Variables["result"] != int64(120) {
		t.Fatalf("factorial result = %v, want 120", runtime.Variables["result"])
	}
}

func TestRuntimeRecursionKeepsIndependentLocals(t *testing.T) {
	runtime := executeTypingSource(t, `func fibonacci(int $n): int {
    ($n <= 1) ? { return $n } : {}
    $left = fibonacci($n - 1)
    $right = fibonacci($n - 2)
    return $left + $right
}
$result = fibonacci(6)`)
	defer runtime.Free()
	if runtime.Variables["result"] != int64(8) {
		t.Fatalf("fibonacci result = %v, want 8", runtime.Variables["result"])
	}
}

func TestRuntimeSupportsMutualRecursion(t *testing.T) {
	runtime := executeTypingSource(t, `func isEven(int $n): bool {
    ($n == 0) ? { return true } : {}
    return isOdd($n - 1)
}
func isOdd(int $n): bool {
    ($n == 0) ? { return false } : {}
    return isEven($n - 1)
}
$result = isEven(10)`)
	defer runtime.Free()
	if runtime.Variables["result"] != true {
		t.Fatalf("mutual recursion result = %v, want true", runtime.Variables["result"])
	}
}

func TestRuntimeSupportsRecursiveMethods(t *testing.T) {
	runtime := executeTypingSource(t, `class Calculator {
    func factorial(int $value): int {
        ($value <= 1) ? { return 1 } : { return $value * $this->factorial($value - 1) }
    }
}
$calculator = new Calculator()
$result = $calculator->factorial(6)`)
	defer runtime.Free()
	if runtime.Variables["result"] != int64(720) {
		t.Fatalf("recursive method result = %v, want 720", runtime.Variables["result"])
	}
}

func TestRuntimeEnforcesDeclaredReturnType(t *testing.T) {
	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "ReturnTypeError" {
			t.Fatalf("expected ReturnTypeError, got %v", recovered)
		}
	}()
	executeTypingSource(t, `func invalid(): int { return "wrong" }
$result = invalid()`)
}

func TestRuntimeStopsUnboundedRecursion(t *testing.T) {
	p := parser.NewParser(parser.NewLexer(`func forever(): int { return forever() }`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	runtime := NewRuntime()
	defer runtime.Free()
	runtime.MaxCallDepth = 24
	runtime.Execute(program)

	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "RecursionLimit" {
			t.Fatalf("expected RecursionLimit, got %v", recovered)
		}
	}()
	runtime.CallMethodEvaluated(runtime.Functions["forever"], nil, nil)
}

func TestRuntimeRejectsConstantReassignment(t *testing.T) {
	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "ConstantAssignment" {
			t.Fatalf("expected ConstantAssignment, got %v", recovered)
		}
	}()
	executeTypingSource(t, `const int $limit = 10
$limit = 20`)
}

func TestRuntimeRejectsConstantPropertyReassignment(t *testing.T) {
	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "ConstantAssignment" {
			t.Fatalf("expected ConstantAssignment, got %v", recovered)
		}
	}()
	executeTypingSource(t, `class Limits {
    const int $maximum = 10
    func change() { $this->maximum = 20 }
}
$limits = new Limits()
$limits->change()`)
}

func TestRuntimeRejectsTypedPropertyChange(t *testing.T) {
	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "PropertyTypeError" {
			t.Fatalf("expected PropertyTypeError, got %v", recovered)
		}
	}()
	executeTypingSource(t, `class Profile {
    int $age = 20
    func invalid() { $this->age = "twenty" }
}
$profile = new Profile()
$profile->invalid()`)
}

func TestRuntimeEnforcesNullableUnion(t *testing.T) {
	runtime := executeTypingSource(t, `int? $count = null
$count = 2`)
	defer runtime.Free()
	if runtime.Variables["count"] != int64(2) || runtime.VarTypes["count"] != "int|null" {
		t.Fatalf("nullable value/type = %v / %q", runtime.Variables["count"], runtime.VarTypes["count"])
	}
}

func TestRuntimeRejectsValueOutsideUnion(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil || !strings.Contains(fmt.Sprint(recovered), "Error de Tipado") {
			t.Fatalf("expected union type error, got %v", recovered)
		}
	}()
	executeTypingSource(t, `int? $count = null
$count = "wrong"`)
}

func TestNamedCallableCannotReadCallerLocal(t *testing.T) {
	defer func() {
		recovered := recover()
		jossError, ok := recovered.(*JossError)
		if !ok || jossError.Type != "UndefinedVariable" {
			t.Fatalf("expected lexical UndefinedVariable, got %v", recovered)
		}
	}()
	executeTypingSource(t, `$secret = "caller-local"
func leak() { return $secret }
$value = leak()`)
}
