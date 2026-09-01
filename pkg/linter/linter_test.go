package linter

import (
	"os"
	"path/filepath"
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

func TestLintPathAnalyzesDirectoryAsOneProject(t *testing.T) {
	project := t.TempDir()
	files := map[string]string{
		"Service.joss": `public class Service {
    public static func value(): string {
        return "ok"
    }
}`,
		"Controller.joss": `public class Controller {
    public func index(): string {
        return Service::value()
    }
}`,
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(project, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	issues, err := NewLinter().LintPath(project)
	if err != nil {
		t.Fatalf("lint directory: %v", err)
	}
	for _, issue := range issues {
		if issue.RuleID == "JOSS-SYM-004" || issue.RuleID == "JOSS-SYM-001" {
			t.Fatalf("cross-file symbol produced a false positive: %s", issue.String())
		}
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
