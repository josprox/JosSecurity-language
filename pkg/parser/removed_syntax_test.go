package parser

import (
	"strings"
	"testing"
)

func TestRemovedCompatibilitySyntaxIsRejected(t *testing.T) {
	cases := []string{
		`function old() {}`,
		`func current() { $callback = function() {} }`,
		`import "other.joss"`,
		`@import "global"`,
		`use Package`,
		`namespace legacy`,
		`Namespace Legacy`,
	}
	for _, source := range cases {
		p := NewParser(NewLexer(source))
		p.ParseProgram()
		if errors := p.Errors(); len(errors) == 0 || !strings.Contains(strings.Join(errors, " "), "sintaxis eliminada") {
			t.Fatalf("%q must report removed syntax, got %v", source, errors)
		}
	}
}

func TestCanonicalFuncAndReturnAnnotationParse(t *testing.T) {
	p := NewParser(NewLexer(`func factorial(int $n): int { return $n }
$callback = func(string $value): string { return $value }`))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) > 0 {
		t.Fatalf("canonical syntax failed: %v", errors)
	}
	method, ok := program.Statements[0].(*MethodStatement)
	if !ok || method.ReturnType.Literal != "int" {
		t.Fatalf("named return annotation was not retained: %#v", program.Statements[0])
	}
}

func TestToolingKeywordsExcludeRemovedSyntax(t *testing.T) {
	joined := " " + strings.Join(KeywordNames(), " ") + " "
	for _, removed := range []string{"function", "import", "@import", "use", "Use", "namespace", "Namespace", "Import"} {
		if strings.Contains(joined, " "+removed+" ") {
			t.Fatalf("removed keyword %q leaked into tooling catalog", removed)
		}
	}
}

func TestParserExposesStructuredDiagnosticRange(t *testing.T) {
	p := NewParser(NewLexer("\n  function old() {}"))
	p.ParseProgram()
	items := p.Diagnostics()
	if len(items) == 0 {
		t.Fatal("expected structured parser diagnostic")
	}
	if items[0].Code != "JOSS-PARSE-001" || items[0].Range.Start.Line != 2 || items[0].Range.Start.Column == 0 {
		t.Fatalf("unexpected parser diagnostic: %#v", items[0])
	}
}

func TestConstantDeclarationsRequireInitializer(t *testing.T) {
	valid := NewParser(NewLexer(`const $limit = 10
const string $name = "Joss"`))
	program := valid.ParseProgram()
	if errors := valid.Errors(); len(errors) > 0 {
		t.Fatalf("constant declarations failed: %v", errors)
	}
	first, ok := program.Statements[0].(*LetStatement)
	if !ok || !first.IsConst || first.Token.Literal != "var" {
		t.Fatalf("inferred constant AST is invalid: %#v", program.Statements[0])
	}

	invalid := NewParser(NewLexer(`const $missing`))
	invalid.ParseProgram()
	if len(invalid.Errors()) == 0 {
		t.Fatal("constant without initializer must fail")
	}
}
