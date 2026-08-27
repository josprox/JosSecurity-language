package core

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

// runSource parses and executes a Joss source string in an isolated runtime.
func runSource(source string) {
	l := parser.NewLexer(source)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	rt := &Runtime{
		Variables:         make(map[string]interface{}),
		VarTypes:          make(map[string]string),
		Constants:         make(map[string]bool),
		Classes:           make(map[string]*parser.ClassStatement),
		Functions:         make(map[string]*parser.MethodStatement),
		Routes:            make(map[string]map[string]interface{}),
		CurrentMiddleware: make([]string, 0),
		CustomMiddlewares: make(map[string]interface{}),
		NativeHandlers:    make(map[string]NativeHandler),
		NativePlugins:     make(map[string]*NativePluginDefinition),
		NativeDrivers:     make(map[string]*NativeDriverDefinition),
		Env:               make(map[string]string),
	}
	rt.Execute(program)
}

// mustPanicWithType asserts that running source panics with a *JossError of the given type.
func mustPanicWithType(t *testing.T, source string, wantType string) {
	t.Helper()
	panicked := false
	var caught interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				caught = r
			}
		}()
		runSource(source)
	}()
	if !panicked {
		t.Errorf("expected panic with JossError{Type:%q} but code ran without error", wantType)
		return
	}
	je, ok := caught.(*JossError)
	if !ok {
		t.Errorf("expected *JossError but got %T: %v", caught, caught)
		return
	}
	if je.Type != wantType {
		t.Errorf("expected JossError.Type=%q, got %q\nMessage: %s", wantType, je.Type, je.Message)
	}
}

// mustNotPanic asserts that running source completes without panic.
func mustNotPanic(t *testing.T, source string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	runSource(source)
}

// ——— Tests ———

func TestUndefinedVariable(t *testing.T) {
	mustPanicWithType(t, `$x = $undefinedVar`, "UndefinedVariable")
}

func TestUndefinedFunction(t *testing.T) {
	mustPanicWithType(t, `nonExistentFunction()`, "UndefinedFunction")
}

func TestUndefinedClass(t *testing.T) {
	mustPanicWithType(t, `$obj = new NonExistentClass()`, "UndefinedClass")
}

func TestIssetDoesNotPanicOnUndefined(t *testing.T) {
	// isset() must return false for undefined vars — must NOT panic
	mustNotPanic(t, `$result = isset($undefinedVar)`)
}

func TestEmptyDoesNotPanicOnUndefined(t *testing.T) {
	// empty() must return true for undefined vars — must NOT panic
	mustNotPanic(t, `$result = empty($undefinedVar)`)
}

func TestDefinedVariableWorks(t *testing.T) {
	mustNotPanic(t, `$x = 42`)
}

func TestDefinedFunctionWorks(t *testing.T) {
	mustNotPanic(t, `
public func greet() {
	return "hello"
}
$result = greet()
`)
}

func TestNullMemberAccessPanics(t *testing.T) {
	// In Joss, member access uses -> (arrow) syntax.
	// Accessing a non-existent property on a real instance → UndefinedProperty
	mustPanicWithType(t, `
public class Foo {}
$f = new Foo()
$v = $f->nonExistentProp
`, "UndefinedProperty")
}

func TestTryCatchCatchesJossError(t *testing.T) {
	// try/catch should catch undefined function errors and expose structured fields
	mustNotPanic(t, `
try {
	nonExistentFunction()
} catch ($e) {
	// $e should be a map with type, message, file fields
}
`)
}

func TestJossErrorFormat(t *testing.T) {
	je := &JossError{
		Type:    "UndefinedVariable",
		Message: "Variable 'foo' no definida",
		File:    "test.joss",
		Line:    10,
	}
	msg := je.Error()
	if msg == "" {
		t.Error("JossError.Error() returned empty string")
	}
	if je.Type != "UndefinedVariable" {
		t.Errorf("unexpected type: %q", je.Type)
	}
}
