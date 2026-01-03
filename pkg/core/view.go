package core

import (
	"fmt"
	"html"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jossecurity/joss/pkg/i18n"
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
	// Inject data into variables
	for k, v := range data {
		// Fix: Don't prepend $ here, as Parser/Evaluator expects raw identifier name
		r.Variables[k] = v
	}

	l := parser.NewLexer(expr)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Sprintf("Error parsing view expression: %s | Details: %v", expr, p.Errors())
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
	// Initialize AssetManager on first use
	am := GetAssetManager()
	am.Initialize()

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
				// fmt.Println("[View DEBUG] Found $__session")
				if sessInst, ok := sessVal.(*Instance); ok {
					// fmt.Printf("[View DEBUG] Session keys: %v\n", sessInst.Fields)
					if _, ok := sessInst.Fields["user_id"]; ok {
						data["auth_check"] = true
						data["auth_guest"] = false
						if name, ok := sessInst.Fields["user_name"]; ok {
							data["auth_user"] = name
						}
						if role, ok := sessInst.Fields["user_role"]; ok {
							data["auth_role"] = role
						}
						if email, ok := sessInst.Fields["user_email"]; ok {
							data["auth_email"] = email
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
						// fmt.Printf("[View DEBUG] Injecting CSRF token: %v\n", csrfVal)
						data["csrf_token"] = csrfVal
					} else {
						fmt.Println("[View DEBUG] CSRF token NOT FOUND in $__session fields")
						// Print all keys for debugging
						for k := range sessInst.Fields {
							fmt.Printf("[View DEBUG] Available Key: %s\n", k)
						}
					}
				} else {
					fmt.Println("[View DEBUG] $__session is not an Instance")
				}
			} else {
				fmt.Println("[View DEBUG] $__session variable NOT FOUND in Runtime")
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
				// Support both ' and " quotes
				reExtends := regexp.MustCompile(`@extends\s*\(\s*['"]([^'"]+)['"]\s*\)`)
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
				// Support both ' and " quotes
				reSection := regexp.MustCompile(`@section\s*\(\s*['"]([^'"]+)['"]\s*\)([\s\S]*?)@endsection`)
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
					// Make placeholder regex-safe or try both quote types
					finalHtml = strings.ReplaceAll(finalHtml, fmt.Sprintf("@yield('%s')", name), content)
					finalHtml = strings.ReplaceAll(finalHtml, fmt.Sprintf("@yield(\"%s\")", name), content)
				}
				// Remove any remaining @yields
				reYield := regexp.MustCompile(`@yield\s*\(\s*['"][^'"]+['"]\s*\)`)
				finalHtml = reYield.ReplaceAllString(finalHtml, "")
			}

			// 3.4 Handle @include('view.name')
			reInclude := regexp.MustCompile(`@include\s*\(\s*['"]([^'"]+)['"]\s*\)`)
			for {
				match := reInclude.FindStringSubmatch(finalHtml)
				if match == nil {
					break
				}
				fullMatch := match[0]
				includeName := match[1]
				includePath := strings.ReplaceAll(includeName, ".", "/")
				var includeContent []byte

				// Resolve Path (reuse logic or simplify)
				// Note: We are repeating file reading logic here. In a full refactor we should extract 'readView(name)' helper.
				// For now, inline to be safe.

				if GlobalFileSystem != nil {
					// VFS
					iPath := path.Join("app", "views", includePath+".joss.html")
					f, err := GlobalFileSystem.Open(iPath)
					if err == nil {
						stat, _ := f.Stat()
						c := make([]byte, stat.Size())
						f.Read(c)
						f.Close()
						includeContent = c
					} else {
						// Try .html
						iPath = path.Join("app", "views", includePath+".html")
						f, err = GlobalFileSystem.Open(iPath)
						if err == nil {
							stat, _ := f.Stat()
							c := make([]byte, stat.Size())
							f.Read(c)
							f.Close()
							includeContent = c
						} else {
							includeContent = []byte(fmt.Sprintf("<!-- Error: Include '%s' not found -->", includeName))
						}
					}
				} else {
					// Disk
					iPath := filepath.Join("app", "views", includePath+".joss.html")
					c, err := os.ReadFile(iPath)
					if err == nil {
						includeContent = c
					} else {
						iPath = filepath.Join("app", "views", includePath+".html")
						c, err := os.ReadFile(iPath)
						if err == nil {
							includeContent = c
						} else {
							includeContent = []byte(fmt.Sprintf("<!-- Error: Include '%s' not found -->", includeName))
						}
					}
				}

				finalHtml = strings.Replace(finalHtml, fullMatch, string(includeContent), 1)
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
						for _, itemMap := range listMap {
							itemHtml := blockContent
							for k, v := range itemMap {
								valStr := fmt.Sprintf("%v", v)

								// 1. Dot notation: {{ $item.key }}
								reDot := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\.%s\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
								itemHtml = reDot.ReplaceAllString(itemHtml, valStr)

								// 2. Bracket notation (single quote): {{ $item['key'] }}
								reBracket1 := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\['%s'\]\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
								itemHtml = reBracket1.ReplaceAllString(itemHtml, valStr)

								// 3. Bracket notation (double quote): {{ $item["key"] }}
								reBracket2 := regexp.MustCompile(fmt.Sprintf(`\{\{\s*\$%s\["%s"\]\s*\}\}`, regexp.QuoteMeta(itemName), regexp.QuoteMeta(k)))
								itemHtml = reBracket2.ReplaceAllString(itemHtml, valStr)
							}
							result += itemHtml
						}
					}
				}
				finalHtml = strings.Replace(finalHtml, fullMatch, result, 1)
			}

			// 3.8 Handle Helpers (Pre-Evaluator)
			tokenVal := ""
			if token, ok := data["csrf_token"]; ok {
				tokenVal = fmt.Sprintf("%v", token)
			}

			// Use Regex for whitespace flexibility
			reCsrf := regexp.MustCompile(`\{\{\s*csrf_field\(\)\s*\}\}`)
			if reCsrf.MatchString(finalHtml) {
				fmt.Printf("[View DEBUG] Replaced {{ csrf_field() }} via Regex with token: %s\n", tokenVal)
				field := fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, tokenVal)
				finalHtml = reCsrf.ReplaceAllString(finalHtml, field)
			} else {
				// Fallback
				if strings.Contains(finalHtml, "csrf_field()") {
					fmt.Println("[View DEBUG] Regex failed but found 'csrf_field()'. Attempting direct replace.")
					field := fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, tokenVal)
					finalHtml = strings.ReplaceAll(finalHtml, "{{ csrf_field() }}", field)
					finalHtml = strings.ReplaceAll(finalHtml, "{{csrf_field()}}", field)
				}
			}

			// Handle I18n helper: {{ __('key') }}
			// Regex to match {{ __('key') }} or {{ __("key") }}
			reLang := regexp.MustCompile(`\{\{\s*__\(\s*['"]([^'"]+)['"]\s*\)\s*\}\}`)
			finalHtml = reLang.ReplaceAllStringFunc(finalHtml, func(match string) string {
				submatch := reLang.FindStringSubmatch(match)
				if len(submatch) > 1 {
					key := submatch[1]
					locale := r.GetLocale() // Use current locale
					return i18n.GlobalManager.Get(locale, key, nil)
				}
				return match
			})

			// 4. Variable Replacement (Evaluator Based)

			// Handle {{ ... }}
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
				if strings.Contains(content, "? {") {
					return match // Skip
				}

				// Evaluate
				val := r.evaluateViewExpression(content, data)

				// Handle nil explicitly to avoid printing "<nil>"
				if val == nil {
					return ""
				}

				valStr := fmt.Sprintf("%v", val)

				if isRaw {
					return valStr
				}
				return html.EscapeString(valStr)
			})

			// 5. Asset Injection (Node Modules Only)
			vendorCSS, vendorJS := am.GetVendorIncludes()

			// REVERT: Dynamic CSS removed by user request.
			// System relies on static <link href="/public/css/app.css"> in layout.

			// Inject CSS (Vendor Only)
			if strings.Contains(finalHtml, "<!-- JOSS_ASSETS -->") {
				// Custom placeholder
				finalHtml = strings.Replace(finalHtml, "<!-- JOSS_ASSETS -->", vendorCSS, 1) // Removed dynamicCSS
			} else if strings.Contains(finalHtml, "</head>") {
				// Inject before head close
				finalHtml = strings.Replace(finalHtml, "</head>", vendorCSS+"</head>", 1) // Removed dynamicCSS
			} else {
				// Just prepend
				finalHtml = vendorCSS + finalHtml // Removed dynamicCSS
			}

			// Inject JS
			if strings.Contains(finalHtml, "</body>") {
				finalHtml = strings.Replace(finalHtml, "</body>", vendorJS+"</body>", 1)
			} else {
				finalHtml = finalHtml + vendorJS
			}

			return finalHtml
		}
	}
	return nil
}
