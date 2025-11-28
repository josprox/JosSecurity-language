package core

import (
	"fmt"
	"os"
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
				}
			}

			// 1. Read View Content
			viewPath := strings.ReplaceAll(viewName, ".", "/")
			path := filepath.Join("app", "views", viewPath+".joss.html")
			content, err := os.ReadFile(path)
			if err != nil {
				path = filepath.Join("app", "views", viewPath+".html")
				content, err = os.ReadFile(path)
				if err != nil {
					return fmt.Sprintf("Error: Vista '%s' no encontrada", viewName)
				}
			}
			viewContent := string(content)

			// 2. Handle Inheritance (@extends)
			var layoutContent string
			sections := make(map[string]string)

			if strings.HasPrefix(strings.TrimSpace(viewContent), "@extends") {
				// Extract layout name
				reExtends := regexp.MustCompile(`@extends\('([^']+)'\)`)
				match := reExtends.FindStringSubmatch(viewContent)
				if len(match) > 1 {
					layoutName := match[1]
					layoutPath := strings.ReplaceAll(layoutName, ".", "/")
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

				// Extract Sections
				// @section('name') ... @endsection
				// We need a loop to find all sections
				reSection := regexp.MustCompile(`@section\('([^']+)'\)([\s\S]*?)@endsection`)
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
				reYield := regexp.MustCompile(`@yield\('[^']+'\)`)
				finalHtml = reYield.ReplaceAllString(finalHtml, "")
			}

			// 4. Variable Replacement

			// A. Handle Ternaries: {{ $var ? 'trueVal' : 'falseVal' }}
			// Regex for: {{ $var ? 'a' : 'b' }} (supporting single quotes)
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

			// B. Handle Null Coalescing: {{ $var ?? "default" }}
			reCoalesce := regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+)\s*\?\?\s*"([^"]*)"\s*\}\}`)
			finalHtml = reCoalesce.ReplaceAllStringFunc(finalHtml, func(match string) string {
				parts := reCoalesce.FindStringSubmatch(match)
				key := parts[1]
				defaultVal := parts[2]
				if val, ok := data[key]; ok {
					return fmt.Sprintf("%v", val)
				}
				return defaultVal
			})

			// C. Handle Standard Replacements
			for k, v := range data {
				// Strict {{key}}
				placeholder := "{{" + k + "}}"
				finalHtml = strings.ReplaceAll(finalHtml, placeholder, fmt.Sprintf("%v", v))
				// Spaced {{ key }}
				placeholderSpace := "{{ " + k + " }}"
				finalHtml = strings.ReplaceAll(finalHtml, placeholderSpace, fmt.Sprintf("%v", v))
				// Dollar var {{ $key }}
				placeholderVar := "{{ $" + k + " }}"
				finalHtml = strings.ReplaceAll(finalHtml, placeholderVar, fmt.Sprintf("%v", v))
			}

			// D. Cleanup remaining {{ $var }} tags (replace with empty string)
			reRemaining := regexp.MustCompile(`\{\{\s*\$[a-zA-Z0-9_]+\s*\}\}`)
			finalHtml = reRemaining.ReplaceAllString(finalHtml, "")

			return finalHtml
		}
	}
	return nil
}
