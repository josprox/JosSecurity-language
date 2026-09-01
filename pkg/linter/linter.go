package linter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/parser"
)

type RuleCategory string

const (
	CategoryCorrectness RuleCategory = "correctness"
	CategoryStyle       RuleCategory = "style"
	CategorySecurity    RuleCategory = "security"
	CategoryPerformance RuleCategory = "performance"
)

type LintIssue struct {
	RuleID      string               `json:"rule_id"`
	Category    RuleCategory         `json:"category"`
	Severity    diagnostics.Severity `json:"severity"`
	File        string               `json:"file"`
	Line        int                  `json:"line"`
	Column      int                  `json:"column"`
	Message     string               `json:"message"`
	Explanation string               `json:"explanation,omitempty"`
	Suggestion  string               `json:"suggestion,omitempty"`
	AutoFixable bool                 `json:"auto_fixable"`
}

func (i LintIssue) String() string {
	loc := i.File
	if i.Line > 0 {
		loc = fmt.Sprintf("%s:%d:%d", i.File, i.Line, i.Column)
	}
	return fmt.Sprintf("[%s] %s %s: %s", i.RuleID, strings.ToUpper(string(i.Severity)), loc, i.Message)
}

type Linter struct {
	files []string
}

func NewLinter(paths ...string) *Linter {
	return &Linter{files: paths}
}

func (l *Linter) LintSource(filename, src string) ([]LintIssue, error) {
	p := parser.NewParser(parser.NewLexer(src))
	prog := p.ParseProgram()

	var issues []LintIssue

	// 1. Parser Errors
	if len(p.Errors()) > 0 {
		for _, errStr := range p.Errors() {
			issues = append(issues, LintIssue{
				RuleID:   "JOSS-SYNTAX-001",
				Category: CategoryCorrectness,
				Severity: diagnostics.SeverityError,
				File:     filename,
				Line:     1,
				Message:  errStr,
			})
		}
		return issues, nil
	}

	// 2. Semantic Analysis integration
	report := core.AnalyzeProgram(prog)
	for _, diag := range report.Diagnostics {
		issues = append(issues, LintIssue{
			RuleID:      diag.Code,
			Category:    CategoryCorrectness,
			Severity:    diag.Severity,
			File:        filename,
			Line:        diag.Range.Start.Line,
			Column:      diag.Range.Start.Column,
			Message:     diag.Message,
			Explanation: diag.Explanation,
			Suggestion:  diag.Suggestion,
			AutoFixable: false,
		})
	}

	// 3. Static AST Lint Rules
	astIssues := checkASTRules(filename, prog, src)
	issues = append(issues, astIssues...)

	return issues, nil
}

func (l *Linter) LintPath(targetPath string) ([]LintIssue, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	var allIssues []LintIssue

	if !info.IsDir() {
		data, err := os.ReadFile(targetPath)
		if err != nil {
			return nil, err
		}
		return l.LintSource(targetPath, string(data))
	}

	err = filepath.WalkDir(targetPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".joss") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			issues, lintErr := l.LintSource(path, string(data))
			if lintErr != nil {
				return lintErr
			}
			allIssues = append(allIssues, issues...)
		}
		return nil
	})

	return allIssues, err
}

func checkASTRules(filename string, prog *parser.Program, src string) []LintIssue {
	var issues []LintIssue
	if prog == nil {
		return issues
	}

	for _, stmt := range prog.Statements {
		switch s := stmt.(type) {
		case *parser.MethodStatement:
			// Rule: Check function naming (camelCase)
			if s.Name != nil && len(s.Name.Value) > 0 {
				first := rune(s.Name.Value[0])
				if first >= 'A' && first <= 'Z' && s.Name.Value != "Init" {
					issues = append(issues, LintIssue{
						RuleID:      "JOSS-LINT-007",
						Category:    CategoryStyle,
						Severity:    diagnostics.SeverityWarning,
						File:        filename,
						Line:        s.Name.Token.Line,
						Column:      s.Name.Token.Column,
						Message:     fmt.Sprintf("La función '%s' debería usar convención camelCase.", s.Name.Value),
						Suggestion:  fmt.Sprintf("Renombra la función a '%s%s'.", strings.ToLower(string(first)), s.Name.Value[1:]),
						AutoFixable: false,
					})
				}
			}
			// Rule: Explicit parameter types
			for _, param := range s.Parameters {
				if param.Type.Literal == "" || param.Type.Type == parser.VAR {
					issues = append(issues, LintIssue{
						RuleID:      "JOSS-LINT-002",
						Category:    CategoryCorrectness,
						Severity:    diagnostics.SeverityError,
						File:        filename,
						Line:        param.Name.Token.Line,
						Column:      param.Name.Token.Column,
						Message:     fmt.Sprintf("El parámetro '$%s' requiere un tipo explícito (o mixed).", param.Name.Value),
						Suggestion:  fmt.Sprintf("Declara el tipo explícito, por ejemplo: 'mixed $%s'.", param.Name.Value),
						AutoFixable: true,
					})
				}
			}
		case *parser.ClassStatement:
			// Rule: Class names must be PascalCase
			if s.Name != nil && len(s.Name.Value) > 0 {
				first := rune(s.Name.Value[0])
				if first >= 'a' && first <= 'z' {
					issues = append(issues, LintIssue{
						RuleID:      "JOSS-LINT-007",
						Category:    CategoryStyle,
						Severity:    diagnostics.SeverityWarning,
						File:        filename,
						Line:        s.Name.Token.Line,
						Column:      s.Name.Token.Column,
						Message:     fmt.Sprintf("La clase '%s' debería usar convención PascalCase.", s.Name.Value),
						Suggestion:  fmt.Sprintf("Renombra la clase a '%s%s'.", strings.ToUpper(string(first)), s.Name.Value[1:]),
						AutoFixable: false,
					})
				}
			}
		}
	}

	// Security: Check for hardcoded API keys or high-entropy credentials
	lines := strings.Split(src, "\n")
	for idx, line := range lines {
		lower := strings.ToLower(line)
		if (strings.Contains(lower, "password = \"") || strings.Contains(lower, "secret = \"") || strings.Contains(lower, "apikey = \"")) && !strings.Contains(line, "System::env") && !strings.Contains(line, "Env::") {
			issues = append(issues, LintIssue{
				RuleID:      "JOSS-SEC-001",
				Category:    CategorySecurity,
				Severity:    diagnostics.SeverityWarning,
				File:        filename,
				Line:        idx + 1,
				Column:      1,
				Message:     "Posible credencial o secreto sensible codificado en duro.",
				Explanation: "Las credenciales no deben almacenarse en el código fuente.",
				Suggestion:  "Usa System::env(\"VARIABLE\") para cargar credenciales desde el entorno.",
				AutoFixable: false,
			})
		}
	}

	return issues
}
