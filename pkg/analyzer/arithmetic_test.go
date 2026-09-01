package analyzer

import (
	"testing"

	"github.com/jossecurity/joss/pkg/diagnostics"
)

func TestAnalyzerDetectsConstantIntegerOverflow(t *testing.T) {
	items := analyzeSource(t, `$value = 9223372036854775807 + 1`, NewEnvironment())
	if !hasCode(items, diagnostics.CodeArithmeticOverflow) {
		t.Fatalf("expected constant overflow diagnostic, got %#v", items)
	}
}

func TestAnalyzerAllowsExactLargeIntegerArithmetic(t *testing.T) {
	items := analyzeSource(t, `$value = 9007199254740993 + 1 echo $value`, NewEnvironment())
	if hasCode(items, diagnostics.CodeArithmeticOverflow) {
		t.Fatalf("exact in-range integer was rejected: %#v", items)
	}
}

func TestAnalyzerDetectsConstantDivisionByZero(t *testing.T) {
	items := analyzeSource(t, `$value = 1 / 0`, NewEnvironment())
	if !hasCode(items, diagnostics.CodeDivisionByZero) {
		t.Fatalf("expected division-by-zero diagnostic, got %#v", items)
	}
}
