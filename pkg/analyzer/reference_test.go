package analyzer

import (
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func analyzeReferenceSource(t *testing.T, source string) []string {
	t.Helper()
	languageParser := parser.NewParser(parser.NewLexer(source))
	program := languageParser.ParseProgram()
	issues := languageParser.Diagnostics()
	if len(issues) != 0 {
		t.Fatalf("parser diagnostics: %+v", issues)
	}
	diagnostics := Analyze([]SourceUnit{{Path: "ref.joss", Program: program}}, NewEnvironment())
	codes := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	return codes
}

func TestReferenceCallIsAccepted(t *testing.T) {
	codes := analyzeReferenceSource(t, `public func increment(ref int $value): int { $value = $value + 1 return $value }
$age = 20
$result = increment(ref $age)`)
	for _, code := range codes {
		if code == "JOSS-REF-001" || code == "JOSS-REF-002" || code == "JOSS-REF-003" || code == "JOSS-REF-004" {
			t.Fatalf("unexpected reference diagnostic: %v", codes)
		}
	}
}

func TestReferenceRejectsMissingMarkerConstantAndDifferentType(t *testing.T) {
	tests := []struct {
		source string
		code   string
	}{
		{`public func increment(ref int $value): int { return $value }
$age = 20
increment($age)`, "JOSS-REF-001"},
		{`public func increment(ref int $value): int { return $value }
const $age = 20
increment(ref $age)`, "JOSS-REF-003"},
		{`public func increment(ref int $value): int { return $value }
float $age = 20.5
increment(ref $age)`, "JOSS-REF-004"},
	}
	for _, test := range tests {
		codes := analyzeReferenceSource(t, test.source)
		found := false
		for _, code := range codes {
			found = found || code == test.code
		}
		if !found {
			t.Fatalf("codes %v do not contain %s", codes, test.code)
		}
	}
}
