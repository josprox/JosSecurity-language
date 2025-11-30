package core

import (
	"fmt"
	"html"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

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

				condition := false

				// Check for equality: $var == 'val'
				if strings.Contains(condStr, "==") {
					parts := strings.Split(condStr, "==")
					if len(parts) == 2 {
						key := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[0]), "$"))
						val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")

						if v, ok := data[key]; ok {
							if fmt.Sprintf("%v", v) == val {
								condition = true
							}
						}
					}
				} else {
					// Simple boolean check: $var
					key := strings.TrimPrefix(condStr, "$")
					if val, ok := data[key]; ok && val != nil && val != false && val != "" && val != 0 {
						condition = true
					}
				}

				if condition {
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

			// 4. Variable Replacement

			// A. Handle Ternaries with Equality: {{ $var == 'val' ? 'true' : 'false' }} or {{ $var.prop == 1 ... }}
			reTernaryEq := regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)?)\s*==\s*['"]?([^'"\s]+)['"]?\s*\?\s*'([^']*)'\s*:\s*'([^']*)'\s*\}\}`)
			finalHtml = reTernaryEq.ReplaceAllStringFunc(finalHtml, func(match string) string {
				parts := reTernaryEq.FindStringSubmatch(match)
				key := parts[1]
				targetVal := parts[2]
				trueVal := parts[3]
				falseVal := parts[4]

				// Resolve value (handle dot notation)
				var currentVal interface{}
				if strings.Contains(key, ".") {
					keyParts := strings.Split(key, ".")
					varName := keyParts[0]
					propName := keyParts[1]
					if val, ok := data[varName]; ok {
						if inst, ok := val.(*Instance); ok {
							if fieldVal, ok := inst.Fields[propName]; ok {
								currentVal = fieldVal
							}
						} else if m, ok := val.(map[string]interface{}); ok {
							if fieldVal, ok := m[propName]; ok {
								currentVal = fieldVal
							}
						}
					}
				} else {
					if val, ok := data[key]; ok {
						currentVal = val
					}
				}

				if currentVal != nil {
					if fmt.Sprintf("%v", currentVal) == targetVal {
						return trueVal
					}
				}
				return falseVal
			})

			// B. Handle Standard Ternaries: {{ $var ? 'trueVal' : 'falseVal' }}
			reTernary := regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+)\s*\?\s*'([^']*)'\s*:\s*'([^']*)'\s*\}\}`)
			finalHtml = reTernary.ReplaceAllStringFunc(finalHtml, func(match string) string {
				parts := reTernary.FindStringSubmatch(match)
				key := parts[1]
				trueVal := parts[2]
				falseVal := parts[3]

				if val, ok := data[key]; ok && val != nil && val != "" && val != false {
					return trueVal
				}
				return falseVal
			})

			// C. Handle Null Coalescing: {{ $var ?? "default" }} or {{ $var ?? 'default' }}
			// Go regexp doesn't support backreferences (\2), so we use alternatives.
			reCoalesce := regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+)\s*\?\?\s*(?:"([^"]*)"|'([^']*)')\s*\}\}`)
			finalHtml = reCoalesce.ReplaceAllStringFunc(finalHtml, func(match string) string {
				parts := reCoalesce.FindStringSubmatch(match)
				key := parts[1]
				// parts[2] is double quoted value, parts[3] is single quoted value.
				// Since they are mutually exclusive alternatives, we can just concatenate them.
				defaultVal := parts[2] + parts[3]

				if val, ok := data[key]; ok && val != nil && val != "" {
					return fmt.Sprintf("%v", val)
				}
				return defaultVal
			})

			// D. Handle Standard Replacements

			// 0. Handle Helpers
			if token, ok := data["csrf_token"]; ok {
				field := fmt.Sprintf(`<input type="hidden" name="_token" value="%v">`, token)
				finalHtml = strings.ReplaceAll(finalHtml, "{{ csrf_field() }}", field)
			}

			// 1. Handle Dot Notation: {{ $var.prop }}
			reDot := regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+)\.([a-zA-Z0-9_]+)\s*\}\}`)
			finalHtml = reDot.ReplaceAllStringFunc(finalHtml, func(match string) string {
				parts := reDot.FindStringSubmatch(match)
				varName := parts[1]
				propName := parts[2]

				if val, ok := data[varName]; ok {
					if inst, ok := val.(*Instance); ok {
						if fieldVal, ok := inst.Fields[propName]; ok {
							return html.EscapeString(fmt.Sprintf("%v", fieldVal))
						}
					} else if m, ok := val.(map[string]interface{}); ok {
						if fieldVal, ok := m[propName]; ok {
							return html.EscapeString(fmt.Sprintf("%v", fieldVal))
						}
					}
				}
				return "" // Return empty if not found
			})

			// 2. Handle Raw Output: {{! var }} or {{!var}}
			for k, v := range data {
				valStr := fmt.Sprintf("%v", v)
				finalHtml = strings.ReplaceAll(finalHtml, "{{! "+k+" }}", valStr)
				finalHtml = strings.ReplaceAll(finalHtml, "{{!"+k+"}}", valStr)
				finalHtml = strings.ReplaceAll(finalHtml, "{{!$"+k+"}}", valStr)
				finalHtml = strings.ReplaceAll(finalHtml, "{{! $"+k+" }}", valStr)

				escapedVal := html.EscapeString(valStr)
				finalHtml = strings.ReplaceAll(finalHtml, "{{"+k+"}}", escapedVal)
				finalHtml = strings.ReplaceAll(finalHtml, "{{ "+k+" }}", escapedVal)
				finalHtml = strings.ReplaceAll(finalHtml, "{{ $"+k+" }}", escapedVal)
			}

			// E. Cleanup remaining {{ $var }} tags
			reRemaining := regexp.MustCompile(`\{\{\s*\$[a-zA-Z0-9_]+\s*\}\}`)
			finalHtml = reRemaining.ReplaceAllString(finalHtml, "")

			return finalHtml
		}
	}
	return nil
}
