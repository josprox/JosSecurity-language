package core

import (
	"fmt"
	"html"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

// Helper to evaluate an expression string within the current runtime context
func (r *Runtime) evaluateViewExpression(expr string, data map[string]interface{}) interface{} {
	// Create a temporary runtime or use current one?
	// We use the current runtime 'r', but we need to inject 'data' into variables temporarily.
	// Or better, we just ensure 'data' is in r.Variables.
	// In executeViewMethod, we should merge 'data' into r.Variables or a scope.
	// Since we don't have scopes easily accessible here without pushing a new environment,
	// let's just use r.Variables but be careful not to pollute global scope permanently if possible.
	// Actually, executeViewMethod is called within a request, so r is already a forked runtime.
	// We can safely modify r.Variables.

	// Inject data into variables
	for k, v := range data {
		r.Variables["$"+k] = v
	}

	l := parser.NewLexer(expr)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Sprintf("Error parsing view expression: %s", expr)
	}

	if len(program.Statements) == 0 {
		return ""
	}

	// Evaluate the first statement (assuming it's an expression)
	// If it's multiple statements, we execute all and return last result?
	var result interface{}
	for _, stmt := range program.Statements {
		if exprStmt, ok := stmt.(*parser.ExpressionStatement); ok {
			result = r.evaluateExpression(exprStmt.Expression)
		} else {
			// Allow other statements? Maybe not for {{ }}
			result = r.executeStatement(stmt)
		}
	}
	return result
}

// View Implementation
func (r *Runtime) executeViewMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "render" {
		if len(args) >= 1 {
			viewName := args[0].(string)
			data := make(map[string]interface{})
			if len(args) > 1 {
				if d, ok := args[1].(map[string]interface{}); ok {
					data = d
				}
			}
			fmt.Printf("[View] Rendering %s with data: %v\n", viewName, data)

			// Inject Global Auth Variables
			data["auth_check"] = false
			data["auth_guest"] = true
			data["auth_user"] = ""
			data["auth_role"] = ""

			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					if _, ok := sessInst.Fields["user_id"]; ok {
						data["auth_check"] = true
						data["auth_guest"] = false
						if name, ok := sessInst.Fields["user_name"]; ok {
							data["auth_user"] = name
						}
						if role, ok := sessInst.Fields["user_role"]; ok {
							data["auth_role"] = role
						}
					}
					// Inject Flash Messages
					if errVal, ok := sessInst.Fields["error"]; ok {
						data["error"] = errVal
						delete(sessInst.Fields, "error")
					}
					if successVal, ok := sessInst.Fields["success"]; ok {
						data["success"] = successVal
						delete(sessInst.Fields, "success")
					}
					// Inject CSRF Token
					if csrfVal, ok := sessInst.Fields["csrf_token"]; ok {
						data["csrf_token"] = csrfVal
					}
				}
			}

			// 1. Read View Content
			viewPath := strings.ReplaceAll(viewName, ".", "/")

			var content []byte
			var err error

			if GlobalFileSystem != nil {
				// VFS Mode
				// Use path.Join for forward slashes
				pathStr := path.Join("app", "views", viewPath+".joss.html")
				f, errOpen := GlobalFileSystem.Open(pathStr)
				if errOpen == nil {
					stat, _ := f.Stat()
					content = make([]byte, stat.Size())
					f.Read(content)
					f.Close()
				} else {
					// Try .html
					pathStr = path.Join("app", "views", viewPath+".html")
					f, errOpen = GlobalFileSystem.Open(pathStr)
					if errOpen == nil {
						stat, _ := f.Stat()
						content = make([]byte, stat.Size())
						f.Read(content)
						f.Close()
					} else {
						return fmt.Sprintf("Error: Vista '%s' no encontrada (VFS)", viewName)
					}
				}
			} else {
				// Disk Mode
				path := filepath.Join("app", "views", viewPath+".joss.html")
				content, err = os.ReadFile(path)
				if err != nil {
					path = filepath.Join("app", "views", viewPath+".html")
					content, err = os.ReadFile(path)
					if err != nil {
						return fmt.Sprintf("Error: Vista '%s' no encontrada", viewName)
					}
				}
			}

			viewContent := string(content)

			// 2. Handle Inheritance (@extends)
			var layoutContent string
			sections := make(map[string]string)

			if strings.HasPrefix(strings.TrimSpace(viewContent), "@extends") {
				// Extract layout name
				// Allow spaces: @extends ( 'layout' )
				reExtends := regexp.MustCompile(`@extends\s*\(\s*'([^']+)'\s*\)`)
				match := reExtends.FindStringSubmatch(viewContent)
				if len(match) > 1 {
					layoutName := match[1]
					layoutPath := strings.ReplaceAll(layoutName, ".", "/")

					if GlobalFileSystem != nil {
						// VFS Layout
						lPath := path.Join("app", "views", layoutPath+".joss.html")
						f, err := GlobalFileSystem.Open(lPath)
						if err == nil {
							stat, _ := f.Stat()
							lContent := make([]byte, stat.Size())
							f.Read(lContent)
							f.Close()
							layoutContent = string(lContent)
						} else {
							// Try .html
							lPath = path.Join("app", "views", layoutPath+".html")
							f, err = GlobalFileSystem.Open(lPath)
							if err == nil {
								stat, _ := f.Stat()
								lContent := make([]byte, stat.Size())
								f.Read(lContent)
								f.Close()
								layoutContent = string(lContent)
							} else {
								return fmt.Sprintf("Error: Layout '%s' no encontrado (VFS)", layoutName)
							}
						}
					} else {
						// Disk Layout
						lPath := filepath.Join("app", "views", layoutPath+".joss.html")
						lContent, err := os.ReadFile(lPath)
						if err == nil {
							layoutContent = string(lContent)
						} else {
							// Try .html
							lPath = filepath.Join("app", "views", layoutPath+".html")
							lContent, err = os.ReadFile(lPath)
							if err == nil {
								layoutContent = string(lContent)
							} else {
								return fmt.Sprintf("Error: Layout '%s' no encontrado", layoutName)
							}
						}
					}
				}

				// Extract Sections
				// @section('name') ... @endsection
				// We need a loop to find all sections
				// Extract Sections
				// @section('name') ... @endsection
				// Allow spaces: @section ( 'name' )
				reSection := regexp.MustCompile(`@section\s*\(\s*'([^']+)'\s*\)([\s\S]*?)@endsection`)
				sectionMatches := reSection.FindAllStringSubmatch(viewContent, -1)
				for _, sm := range sectionMatches {
					sections[sm[1]] = sm[2]
				}
			}

			// 3. Merge Layout and View
			finalHtml := viewContent
			if layoutContent != "" {
				finalHtml = layoutContent
				// Replace @yield('name') with section content
				for name, content := range sections {
					placeholder := fmt.Sprintf("@yield('%s')", name)
					finalHtml = strings.ReplaceAll(finalHtml, placeholder, content)
				}
				// Remove any remaining @yields
				reYield := regexp.MustCompile(`@yield\s*\(\s*'[^']+'\s*\)`)
				finalHtml = reYield.ReplaceAllString(finalHtml, "")
			}

			// 3.5 Handle Control Structures

			// A. Handle Block Ternaries: {{ ($var) ? { trueBlock } : { falseBlock } }}
			reBlockTernary := regexp.MustCompile(`\{\{\s*\((.*?)\)\s*\?\s*\{([\s\S]*?)\}\s*:\s*\{([\s\S]*?)\}\s*\}\}`)
			for {
				match := reBlockTernary.FindStringSubmatch(finalHtml)
				if match == nil {
					break
				}
				fullMatch := match[0]
				condStr := strings.TrimSpace(match[1])
				trueBlock := match[2]
				falseBlock := match[3]

				// Evaluate condition using the evaluator
				condVal := r.evaluateViewExpression(condStr, data)

				if isTruthy(condVal) {
					finalHtml = strings.Replace(finalHtml, fullMatch, trueBlock, 1)
				} else {
					finalHtml = strings.Replace(finalHtml, fullMatch, falseBlock, 1)
				}
			}

			// B. Handle @foreach($list as $item) ... @endforeach
			// Allow spaces: @foreach ( $list as $item )
			reForeach := regexp.MustCompile(`@foreach\s*\(\s*\$([a-zA-Z0-9_]+)\s+as\s+\$([a-zA-Z0-9_]+)\s*\)([\s\S]*?)@endforeach`)
			for {
				match := reForeach.FindStringSubmatch(finalHtml)
				if match == nil {
					break
				}
				fullMatch := match[0]
				listName := match[1]
				itemName := match[2]
				blockContent := match[3]

				var result string
				if listVal, ok := data[listName]; ok {
					// Handle []interface{} (from script)
					if list, ok := listVal.([]interface{}); ok {
						for _, item := range list {
							itemHtml := blockContent
							if itemMap, ok := item.(map[string]interface{}); ok {
								for k, v := range itemMap {
									valStr := fmt.Sprintf("%v", v)
									// Use regex to replace {{ $item.key }} with flexibility for spaces
									// Pattern: {{ \s* $itemName \. key \s* }}
									reItem := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\.%s\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
									itemHtml = reItem.ReplaceAllString(itemHtml, valStr)
								}
							} else if itemInst, ok := item.(*Instance); ok {
								for k, v := range itemInst.Fields {
									valStr := fmt.Sprintf("%v", v)
									// Use regex to replace {{ $item.key }} with flexibility for spaces
									reItem := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\.%s\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
									itemHtml = reItem.ReplaceAllString(itemHtml, valStr)
								}
							}
							result += itemHtml
						}
					} else if listMap, ok := listVal.([]map[string]interface{}); ok {
						// Handle []map[string]interface{} (from GranDB)
						for _, itemMap := range listMap {
							itemHtml := blockContent
							for k, v := range itemMap {
								valStr := fmt.Sprintf("%v", v)
								// Use regex to replace {{ $item.key }} with flexibility for spaces
								reItem := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\.%s\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
								itemHtml = reItem.ReplaceAllString(itemHtml, valStr)
							}
							result += itemHtml
						}
					}
				}
				finalHtml = strings.Replace(finalHtml, fullMatch, result, 1)
			}

			// 3.8 Handle Helpers (Pre-Evaluator)
			if token, ok := data["csrf_token"]; ok {
				field := fmt.Sprintf(`<input type="hidden" name="_token" value="%v">`, token)
				finalHtml = strings.ReplaceAll(finalHtml, "{{ csrf_field() }}", field)
			}

			// 4. Variable Replacement (Evaluator Based)

			// Handle {{ ... }}
			// We use a loop to find all {{ ... }} blocks that are NOT {{! ... }} (raw) or control structures already handled.
			// Actually, we should handle {{! ... }} first to avoid confusion, or handle them together.

			// Regex to find {{ ... }}
			reTags := regexp.MustCompile(`\{\{(.*?)\}\}`)
			finalHtml = reTags.ReplaceAllStringFunc(finalHtml, func(match string) string {
				content := match[2 : len(match)-2] // Remove {{ and }}
				content = strings.TrimSpace(content)

				// Check for Raw Output {{! ... }}
				isRaw := false
				if strings.HasPrefix(content, "!") {
					isRaw = true
					content = strings.TrimSpace(strings.TrimPrefix(content, "!"))
				}

				// Skip if it looks like a block ternary start/end or other control structure leftovers
				// (Though block ternaries should be handled by now)
				if strings.Contains(content, "? {") {
					return match // Skip, let block ternary handler deal with it (if any left)
				}

				// Evaluate
				val := r.evaluateViewExpression(content, data)
				valStr := fmt.Sprintf("%v", val)

				if isRaw {
					return valStr
				}
				return html.EscapeString(valStr)
			})

			return finalHtml
		}
	}
	return nil
}
