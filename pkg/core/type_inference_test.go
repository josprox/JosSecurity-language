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
