package analyzer

import (
	"testing"

	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

func analyzeSource(t *testing.T, source string, environment Environment) []diagnostics.Diagnostic {
	t.Helper()
	p := parser.NewParser(parser.NewLexer(source))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("parse errors: %v", errors)
	}
	return Analyze([]SourceUnit{{Path: "test.joss", Program: program}}, environment)
}

func hasCode(items []diagnostics.Diagnostic, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}

func TestInferredVariableKeepsFirstConcreteType(t *testing.T) {
	items := analyzeSource(t, `$age = 20
$age = 30
$age = "twenty"`, NewEnvironment())
	if !hasCode(items, "JOSS-TYPE-001") {
		t.Fatalf("expected inferred assignment error, got %#v", items)
	}
}

func TestExplicitDynamicVariableMayChangeType(t *testing.T) {
	items := analyzeSource(t, `let $value = 20
$value = "twenty"
echo $value`, NewEnvironment())
	if hasCode(items, "JOSS-TYPE-001") || hasCode(items, "JOSS-TYPE-002") {
		t.Fatalf("mixed variable should stay dynamic, got %#v", items)
	}
}

func TestVarDeclarationInfersAndTypedDeclarationChecks(t *testing.T) {
	items := analyzeSource(t, `var $age = 20
$age = "twenty"
string $name = 42`, NewEnvironment())
	if !hasCode(items, "JOSS-TYPE-001") || !hasCode(items, "JOSS-TYPE-002") {
		t.Fatalf("expected assignment and initializer diagnostics, got %#v", items)
	}
}

func TestTypedNumericStringLiteralUsesRuntimeCoercionPolicy(t *testing.T) {
	items := analyzeSource(t, `int $age = "20"
echo $age`, NewEnvironment())
	if hasCode(items, "JOSS-TYPE-002") {
		t.Fatalf("lossless numeric string should be accepted consistently: %#v", items)
	}
}

func TestCallableScopesDoNotLeak(t *testing.T) {
	items := analyzeSource(t, `class Example {
  func first($value) { $local = $value echo $local }
  func second($value) { $local = $value echo $local }
}`, NewEnvironment())
	if hasCode(items, "JOSS-SYM-002") || hasCode(items, "JOSS-SYM-001") {
		t.Fatalf("method-local symbols leaked across scopes: %#v", items)
	}
}

func TestTypedFunctionArgumentsAndArity(t *testing.T) {
	items := analyzeSource(t, `func add(int $a, int $b) { return $a + $b }
add(1, "two")
add(1)`, NewEnvironment())
	if !hasCode(items, "JOSS-TYPE-003") || !hasCode(items, "JOSS-CALL-001") {
		t.Fatalf("expected argument type and arity diagnostics, got %#v", items)
	}
}

func TestUnknownNativeSignatureDoesNotCreateArityFalsePositive(t *testing.T) {
	environment := NewEnvironment()
	environment.Classes["Native"] = Class{Name: "Native", Methods: map[string]Callable{
		"call": {Name: "call", Variadic: true, ReturnType: typesystem.Type{Kind: typesystem.Unknown}},
	}}
	items := analyzeSource(t, `Native::call(1, 2, 3)`, environment)
	if hasCode(items, "JOSS-CALL-001") {
		t.Fatalf("unknown native signature must not report arity: %#v", items)
	}
}

func TestIssetDoesNotTreatUnknownAsInvalid(t *testing.T) {
	items := analyzeSource(t, `isset($optional)`, NewEnvironment())
	if hasCode(items, "JOSS-SYM-001") {
		t.Fatalf("isset is an existence probe, got %#v", items)
	}
}

func TestForeachBindingCanBeReused(t *testing.T) {
	items := analyzeSource(t, `$one = []
$two = []
foreach ($one as $item) { echo $item }
foreach ($two as $item) { echo $item }`, NewEnvironment())
	if hasCode(items, "JOSS-SYM-002") {
		t.Fatalf("foreach reuse should be assignment-like, got %#v", items)
	}
}

func TestLocalReceiverShadowsSameNamedClassAndCountsAsUse(t *testing.T) {
	items := analyzeSource(t, `class repository {
  func first() { return 1 }
}
func load() {
  $repository = new repository()
  return $repository->first()
}`, NewEnvironment())
	if hasCode(items, "JOSS-LINT-001") || hasCode(items, "JOSS-MEMBER-001") {
		t.Fatalf("instance receiver was mistaken for its same-named class: %#v", items)
	}
}

func TestDiagnosticsRetainFileAndPosition(t *testing.T) {
	items := analyzeSource(t, "\n echo $missing", NewEnvironment())
	if len(items) == 0 || items[0].File != "test.joss" || items[0].Range.Start.Line != 2 || items[0].Range.Start.Column == 0 {
		t.Fatalf("missing source location: %#v", items)
	}
}

func TestUnreachableStatementIsWarning(t *testing.T) {
	items := analyzeSource(t, `func stop() { return 1 echo "never" }`, NewEnvironment())
	if !hasCode(items, "JOSS-FLOW-001") {
		t.Fatalf("expected unreachable warning, got %#v", items)
	}
}
