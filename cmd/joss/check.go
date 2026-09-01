package main

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/formatter"
	"github.com/jossecurity/joss/pkg/linter"
)

func handleCheckCommand(args []string) {
	var targetPath string
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		targetPath = "."
	}

	fmt.Printf("[CHECK] Verificando proyecto en '%s' (parse, semantic analysis, types, lint, format)...\n\n", targetPath)

	hasErrors := false

	// 1. Format check
	unformatted, err := formatter.FormatDirectory(targetPath, false, true)
	if err != nil {
		fmt.Printf("[CHECK] Error escaneando formato: %v\n", err)
		hasErrors = true
	} else if len(unformatted) > 0 {
		fmt.Printf("[CHECK WARNING] %d archivo(s) no cumplen con el formato canónico (ejecuta 'joss format --write'):\n", len(unformatted))
		for _, f := range unformatted {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
	} else {
		fmt.Println("  ✓ Formato: OK")
	}

	// 2. Lint & Semantic Analysis check
	l := linter.NewLinter()
	issues, err := l.LintPath(targetPath)
	if err != nil {
		fmt.Printf("[CHECK ERROR] Falló el análisis de lint: %v\n", err)
		hasErrors = true
	} else {
		errCount := 0
		warnCount := 0
		for _, issue := range issues {
			if issue.Severity == diagnostics.SeverityError {
				errCount++
				hasErrors = true
			} else {
				warnCount++
			}
		}

		if len(issues) > 0 {
			fmt.Printf("\n[CHECK DIAGNÓSTICOS] %d error(es), %d advertencia(s):\n", errCount, warnCount)
			for _, issue := range issues {
				fmt.Printf("  %s\n", issue.String())
				if issue.Suggestion != "" {
					fmt.Printf("    sugerencia: %s\n", issue.Suggestion)
				}
			}
		} else {
			fmt.Println("  ✓ Análisis semántico y tipos: OK")
			fmt.Println("  ✓ Reglas de linter y seguridad: OK")
		}
	}

	fmt.Println()
	if hasErrors {
		fmt.Println("[CHECK RESULT] ❌ El proyecto contiene problemas que deben resolverse.")
		os.Exit(1)
	}

	fmt.Println("[CHECK RESULT]  El proyecto está completamente verificado y listo.")
}
