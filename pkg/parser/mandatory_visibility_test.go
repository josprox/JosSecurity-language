package parser

import "testing"

func TestNamedClassesFunctionsMethodsAndPropertiesRequireVisibility(t *testing.T) {
	sources := []string{
		`class Missing {}`,
		`func missing() {}`,
		`public class Missing { func method() {} }`,
		`public class Missing { int $value = 1 }`,
		`public class Missing { const int $value = 1 }`,
	}
	for _, source := range sources {
		languageParser := NewParser(NewLexer(source))
		languageParser.ParseProgram()
		if len(languageParser.Diagnostics()) == 0 {
			t.Fatalf("expected visibility diagnostic for %q", source)
		}
	}
}

func TestExplicitVisibilityOnConstantPropertyIsPreserved(t *testing.T) {
	languageParser := NewParser(NewLexer(`public class Limits { private const int $maximum = 10 }`))
	program := languageParser.ParseProgram()
	if issues := languageParser.Diagnostics(); len(issues) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", issues)
	}
	class := program.Statements[0].(*ClassStatement)
	property := class.Body.Statements[0].(*LetStatement)
	if property.Visibility != "private" || !property.IsConst {
		t.Fatalf("constant property = %#v", property)
	}
}

func TestExplicitVisibilityAndClosuresRemainValid(t *testing.T) {
	source := `public class Service {
    private int $value = 1
    protected func value(): int { return $this->value }
}
public func run(): int {
    $callback = func (): int { return 1 }
    return $callback()
}`
	languageParser := NewParser(NewLexer(source))
	languageParser.ParseProgram()
	if issues := languageParser.Diagnostics(); len(issues) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", issues)
	}
}
