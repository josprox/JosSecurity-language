package server

import "testing"

func TestReplaceSCSSVariablesKeepsPrefixNamesDistinct(t *testing.T) {
	variables := map[string]string{
		"$primary":      "#00f2ff",
		"$primary-dark": "#00bceb",
		"$text":         "#f1f5f9",
		"$text-dim":     "#64748b",
	}
	source := ".button { color: $text-dim; background: $primary-dark; border-color: $missing; }"
	want := ".button { color: #64748b; background: #00bceb; border-color: $missing; }"

	if got := replaceSCSSVariables(source, variables); got != want {
		t.Fatalf("replaceSCSSVariables() = %q, want %q", got, want)
	}
}
