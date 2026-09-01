package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jossecurity/joss/pkg/formatter"
	"github.com/jossecurity/joss/pkg/linter"
)

func TestToolingFormatAndLintEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "sample.joss")

	unformattedCode := `public   func   calculate( int $x,int $y ):int{return $x+$y;}`
	if err := os.WriteFile(sourceFile, []byte(unformattedCode), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Check formatter detected unformatted file
	changed, err := formatter.FormatFile(sourceFile, false)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !changed {
		t.Fatalf("expected format to detect changes")
	}

	// 2. Format with write = true
	changed, err = formatter.FormatFile(sourceFile, true)
	if err != nil {
		t.Fatalf("format write error: %v", err)
	}
	if !changed {
		t.Fatalf("expected format to write changes")
	}

	data, _ := os.ReadFile(sourceFile)
	if !strings.Contains(string(data), "public func calculate(int $x, int $y): int {") {
		t.Fatalf("file not properly formatted: %s", string(data))
	}

	// 3. Lint the formatted file
	l := linter.NewLinter()
	issues, err := l.LintPath(sourceFile)
	if err != nil {
		t.Fatalf("lint path error: %v", err)
	}

	// Calculate function has clean types and camelCase naming, so no error issues
	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Fatalf("unexpected error issue in formatted code: %v", issue)
		}
	}
}
