package fixer

import (
	"strings"
	"testing"
)

func TestFixerAddsVisibilityAndFormats(t *testing.T) {
	input := `func compute(int $x): int {
return $x+1;
}`
	fixer := NewFixer(true)
	fixed, count := fixer.FixSource(input)

	if count == 0 {
		t.Fatalf("expected fixes to be applied")
	}
	if !strings.Contains(fixed, "public func compute(int $x): int {") {
		t.Fatalf("expected visibility to be added, got:\n%s", fixed)
	}
	if !strings.Contains(fixed, "    return $x + 1;") {
		t.Fatalf("expected indentation and spacing to be formatted, got:\n%s", fixed)
	}
}
