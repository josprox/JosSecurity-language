package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TranslationItem struct {
	Key     string `json:"key"`
	Spanish string `json:"es"`
	File    string `json:"file"`
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	results := make(map[string]TranslationItem)
	rawMap := make(map[string]string)
	arbMap := make(map[string]interface{})

	// Target function calls for user-facing output
	targetFuncs := map[string]bool{
		"Println":       true,
		"Printf":        true,
		"Print":         true,
		"Errorf":        true,
		"New":           true,
		"ReadString":    true,
		"WriteFile":     true,
		"Write":         true,
		"Fatal":         true,
		"Fatalf":        true,
		"Fatalln":       true,
	}

	reClean := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") || strings.Contains(path, "tools") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funcName := ""
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				funcName = fun.Sel.Name
			case *ast.Ident:
				funcName = fun.Name
			}

			if !targetFuncs[funcName] {
				return true
			}

			for _, arg := range call.Args {
				lit, ok := arg.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}

				val := strings.Trim(lit.Value, "`\"")
				valTrim := strings.TrimSpace(val)
				if len(valTrim) < 3 {
					continue
				}

				// Skip code fragments, SQL, format specifiers, paths, or pure symbols
				if !isTranslatableSpanishText(valTrim) {
					continue
				}

				// Generate camelCase key
				clean := reClean.ReplaceAllString(valTrim, " ")
				words := strings.Fields(clean)
				if len(words) == 0 {
					continue
				}

				keyName := ""
				for i, w := range words {
					if i >= 6 {
						break
					}
					wLower := strings.ToLower(w)
					if i == 0 {
						keyName += wLower
					} else if len(wLower) > 0 {
						keyName += strings.ToUpper(wLower[:1]) + wLower[1:]
					}
				}

				relPath, _ := filepath.Rel(rootDir, path)
				relPath = filepath.ToSlash(relPath)

				if _, exists := results[valTrim]; !exists {
					results[valTrim] = TranslationItem{
						Key:     keyName,
						Spanish: valTrim,
						File:    relPath,
					}
					rawMap[keyName] = valTrim
					arbMap[keyName] = valTrim
					arbMap["@"+keyName] = map[string]string{"description": "Extracted from " + relPath}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning files: %v\n", err)
		os.Exit(1)
	}

	// Format 1: Array of items
	var itemList []TranslationItem
	for _, item := range results {
		itemList = append(itemList, item)
	}
	sort.Slice(itemList, func(i, j int) bool {
		return itemList[i].Key < itemList[j].Key
	})

	jsonItems, _ := json.MarshalIndent(itemList, "", "  ")

	// Format 2: Direct Key-Value Dictionary (Standard JSON)
	jsonKeyValue, _ := json.MarshalIndent(rawMap, "", "  ")

	// Format 3: ARB format (Flutter/i18n ready)
	arbMap["@@locale"] = "es"
	jsonARB, _ := json.MarshalIndent(arbMap, "", "  ")

	// Save to project root
	os.WriteFile("hardcoded_strings_list.json", jsonItems, 0644)
	os.WriteFile("hardcoded_strings.json", jsonKeyValue, 0644)
	os.WriteFile("hardcoded_strings.arb", jsonARB, 0644)

	// Save to sistema-Joss-Red-All if directory exists
	targetDir := `C:\Users\joss\Documents\proyectos\sistema-Joss-Red-All`
	if _, err := os.Stat(targetDir); err == nil {
		os.WriteFile(filepath.Join(targetDir, "hardcoded_strings.json"), jsonKeyValue, 0644)
		os.WriteFile(filepath.Join(targetDir, "hardcoded_strings_list.json"), jsonItems, 0644)
		os.WriteFile(filepath.Join(targetDir, "hardcoded_strings.arb"), jsonARB, 0644)
		fmt.Printf("Copiados archivos JSON traducibles a %s\n", targetDir)
	}

	fmt.Printf("Extraídos %d textos hardcodeados con éxito.\n", len(itemList))
}

func isTranslatableSpanishText(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return false
	}

	// Skip URLs, SQL queries, HTTP methods
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return false
	}
	if strings.HasPrefix(s, "SELECT ") || strings.HasPrefix(s, "CREATE ") || strings.HasPrefix(s, "INSERT ") || strings.HasPrefix(s, "UPDATE ") || strings.HasPrefix(s, "DELETE ") {
		return false
	}

	// Skip pure format specifiers (e.g., "%-30s | %-19s | %-12s")
	clean := regexp.MustCompile(`%[-+0-9.*]*[vdsftxXqT%]`).ReplaceAllString(s, "")
	clean = strings.Trim(clean, " \t\r\n|:;,-=/()[]{}<>.")
	if len(clean) < 2 {
		return false
	}

	// Must contain letters
	hasLetter := false
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127 {
			hasLetter = true
			break
		}
	}
	return hasLetter
}
