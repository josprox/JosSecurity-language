package core

import (
	"strings"
	"testing"

	"github.com/jossecurity/joss/pkg/parser"
)

func TestSyntaxGuardUnsupportedKeywords(t *testing.T) {
	tests := []struct {
		code             string
		expectedContains string
	}{
		{`if ($x > 0) { echo "hola" }`, "La estructura 'if' no existe en Joss"},
		{`else { echo "adios" }`, "La estructura 'else' no existe en Joss"},
		{`elif ($x == 0) { echo "cero" }`, "La estructura 'elif' no existe en Joss"},
		{`switch ($x) { case 1: echo "uno" }`, "La estructura 'switch' no existe en Joss"},
		{`for ($i = 0; $i < 10; $i++) { echo $i }`, "El bucle 'for' no existe en Joss"},
	}

	for _, tt := range tests {
		l := parser.NewLexer(tt.code)
		p := parser.NewParser(l)
		p.ParseProgram()
		errors := p.Errors()
		if len(errors) == 0 {
			t.Errorf("Expected syntax error for code %q, got none", tt.code)
			continue
		}
		found := false
		for _, err := range errors {
			if strings.Contains(err, tt.expectedContains) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error containing %q for code %q, got: %v", tt.expectedContains, tt.code, errors)
		}
	}
}

func TestStaticAnalyzer(t *testing.T) {
	code := `
string $usada = "hola"
string $sinUsar = "chao"
echo $usada

func funcionInexistente() {
	echo $noDeclarada
}
`
	l := parser.NewLexer(code)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	report := AnalyzeProgram(program)

	if !report.HasIssues() {
		t.Errorf("Expected static analysis issues, got none")
	}

	// Should warn about $sinUsar being unused using the stable diagnostic code.
	hasUnusedVarWarning := false
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == "JOSS-LINT-001" && strings.Contains(diagnostic.Message, "$sinUsar") {
			hasUnusedVarWarning = true
		}
	}
	if !hasUnusedVarWarning {
		t.Errorf("Expected warning for unused variable '$sinUsar', got diagnostics: %v", report.Diagnostics)
	}

	// Should report error for $noDeclarada.
	hasUndeclaredVarError := false
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == "JOSS-SYM-001" && strings.Contains(diagnostic.Message, "$noDeclarada") {
			hasUndeclaredVarError = true
		}
	}
	if !hasUndeclaredVarError {
		t.Errorf("Expected error for undeclared variable '$noDeclarada', got diagnostics: %v", report.Diagnostics)
	}
}

func TestStaticAnalyzerRecognizesImplementedGranDBMethods(t *testing.T) {
	code := `
$categories = GranDB::table("products")->distinct()->pluck("category")
GranDB::transaction(func($db) {
	return GranDB::table("products")->count()
})
`
	program := parser.NewParser(parser.NewLexer(code)).ParseProgram()
	report := AnalyzeProgram(program)

	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Code == "JOSS-MEMBER-001" {
			t.Fatalf("implemented GranDB API produced an unknown-member diagnostic: %v", report.Diagnostics)
		}
	}
}
