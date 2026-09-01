package linter

import (
	"testing"

	"github.com/jossecurity/joss/pkg/diagnostics"
)

func TestLinterDetectsUntypedParams(t *testing.T) {
	src := `public func compute($a, int $b): int {
    return $b;
}
`
	linter := NewLinter()
	issues, err := linter.LintSource("test.joss", src)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}

	found := false
	for _, issue := range issues {
		if issue.RuleID == "JOSS-LINT-002" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected JOSS-LINT-002 for untyped param, got: %v", issues)
	}
}

func TestLinterDetectsHardcodedSecret(t *testing.T) {
	src := `public func connect() {
    $apiKey = "secret_key_12345"
}
`
	linter := NewLinter()
	issues, err := linter.LintSource("test.joss", src)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}

	found := false
	for _, issue := range issues {
		if issue.RuleID == "JOSS-SEC-001" {
			found = true
			if issue.Severity != diagnostics.SeverityWarning {
				t.Fatalf("expected warning severity, got %v", issue.Severity)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected JOSS-SEC-001 for hardcoded secret, got: %v", issues)
	}
}
