package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/parser"
)

var parserLinePattern = regexp.MustCompile(`(?i)(?:l[ií]nea|line)\s+(\d+)`)

// LoadProject parses an entrypoint and every .joss file below sourceDirs.
// Paths are retained per AST so later diagnostics never lose file identity.
func LoadProject(entrypoint string, sourceDirs ...string) ([]SourceUnit, []diagnostics.Diagnostic) {
	paths := []string{entrypoint}
	seen := map[string]bool{}
	if absolute, err := filepath.Abs(entrypoint); err == nil {
		seen[filepath.Clean(absolute)] = true
	}

	for _, sourceDir := range sourceDirs {
		info, err := os.Stat(sourceDir)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info == nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if !strings.EqualFold(filepath.Ext(path), ".joss") {
				return nil
			}
			absolute, err := filepath.Abs(path)
			if err == nil && seen[filepath.Clean(absolute)] {
				return nil
			}
			if err == nil {
				seen[filepath.Clean(absolute)] = true
			}
			paths = append(paths, path)
			return nil
		})
	}

	if len(paths) > 1 {
		sort.Strings(paths[1:])
	}
	units := make([]SourceUnit, 0, len(paths))
	issues := make([]diagnostics.Diagnostic, 0)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			issues = append(issues, diagnostics.Diagnostic{
				Code: "JOSS-IO-001", Severity: diagnostics.SeverityError,
				File: path, Message: fmt.Sprintf("Cannot read source file: %v", err),
			})
			continue
		}
		p := parser.NewParser(parser.NewLexer(string(data)))
		program := p.ParseProgram()
		if parseErrors := p.Errors(); len(parseErrors) > 0 {
			for _, parseError := range parseErrors {
				line := 0
				if match := parserLinePattern.FindStringSubmatch(parseError); len(match) == 2 {
					line, _ = strconv.Atoi(match[1])
				}
				issues = append(issues, diagnostics.Diagnostic{
					Code: "JOSS-PARSE-001", Severity: diagnostics.SeverityError,
					File: path, Range: diagnostics.Range{Start: diagnostics.Position{Line: line}},
					Message: parseError,
				})
			}
			continue
		}
		units = append(units, SourceUnit{Path: path, Program: program})
	}
	return units, issues
}
