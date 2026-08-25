package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	semanticanalyzer "github.com/jossecurity/joss/pkg/analyzer"
)

func TestAnalyzerProjectFixtureHasNoErrors(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	project := filepath.Join(root, "testdata", "analyzer-project")

	units, parseDiagnostics := semanticanalyzer.LoadProject(filepath.Join(project, "main.joss"), filepath.Join(project, "app"))
	if len(parseDiagnostics) != 0 {
		t.Fatalf("analyzer fixture parse diagnostics: %#v", parseDiagnostics)
	}
	report := AnalyzeSourceUnits(units)
	if report.HasErrors() {
		t.Fatalf("analyzer fixture analysis errors: %#v", report.Diagnostics)
	}
}

func TestJosSecurityProjectHasNoStaticAnalysisErrors(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	project := filepath.Join(root, "ejemplos", "Joss-Red-JosSecurity")
	if _, err := os.Stat(filepath.Join(project, "main.joss")); err != nil {
		t.Skip("JosSecurity is an external, gitignored integration repository")
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	units, parseDiagnostics := semanticanalyzer.LoadProject("main.joss", "app")
	if len(parseDiagnostics) != 0 {
		t.Fatalf("JosSecurity parse diagnostics: %#v", parseDiagnostics)
	}
	report := AnalyzeSourceUnits(units)
	if report.HasErrors() {
		t.Fatalf("JosSecurity analysis errors: %#v", report.Diagnostics)
	}
}
