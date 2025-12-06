package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

// Dispatch handles HTTP requests based on defined routes
func (r *Runtime) Dispatch(method, path string, reqData, sessData map[string]interface{}) (interface{}, error) {
	// Inject Request and Session
	r.Variables["$__request"] = &Instance{Fields: reqData}
	r.Variables["$__session"] = &Instance{Fields: sessData}

	// Match Route
	var handler interface{}
	var middleware []string
	method = strings.ToUpper(method)

	// Check exact match
	if r.Routes[method] != nil {
		if routeInfo, ok := r.Routes[method][path].(map[string]interface{}); ok {
			handler = routeInfo["handler"]
			if mw, ok := routeInfo["middleware"].([]string); ok {
				middleware = mw
			}
		} else if h, ok := r.Routes[method][path]; ok {
			// Legacy support (just handler)
			handler = h
		}
	}

	// Check dynamic routes if no exact match
	if handler == nil && r.Routes[method] != nil {
		for routePath, routeVal := range r.Routes[method] {
			if strings.Contains(routePath, "{") {
				// Regex matching (simplified)
				regexPath := "^" + regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`).ReplaceAllString(routePath, "([^/]+)") + "$"
				re := regexp.MustCompile(regexPath)
				matches := re.FindStringSubmatch(path)

				if len(matches) > 0 {
					if routeInfo, ok := routeVal.(map[string]interface{}); ok {
						handler = routeInfo["handler"]
						if mw, ok := routeInfo["middleware"].([]string); ok {
							middleware = mw
						}
					} else {
						handler = routeVal
					}

					// Extract params (simplified)
					// In a real implementation, we would map param names to values here
					break
				}
			}
		}
	}

	if handler == nil {
		fmt.Printf("[DEBUG] Route not found: %s %s\n", method, path)
		fmt.Printf("[DEBUG] Available Routes for %s:\n", method)
		if r.Routes[method] != nil {
			for k := range r.Routes[method] {
				fmt.Printf("\t%s\n", k)
			}
		} else {
			fmt.Println("\t(None)")
		}
		return nil, fmt.Errorf("route not found: %s %s", method, path)
	}

	// Middleware Execution
	for _, mw := range middleware {
		switch mw {
		case "auth":
			// Check if logged in
			isLoggedIn := false
			if sessInst, ok := r.Variables["$__session"].(*Instance); ok {
				if _, ok := sessInst.Fields["user_id"]; ok {
					isLoggedIn = true
				}
			}
			if !isLoggedIn {
				return &Instance{
					Fields: map[string]interface{}{
						"_type": "REDIRECT",
						"url":   "/login",
						"flash": map[string]interface{}{
							"error": "Debes iniciar sesión.",
						},
					},
				}, nil
			}
		case "guest":
			// Check if NOT logged in
			isLoggedIn := false
			if sessInst, ok := r.Variables["$__session"].(*Instance); ok {
				if _, ok := sessInst.Fields["user_id"]; ok {
					isLoggedIn = true
				}
			}
			if isLoggedIn {
				return &Instance{
					Fields: map[string]interface{}{
						"_type": "REDIRECT",
						"url":   "/dashboard",
					},
				}, nil
			}
		case "admin":
			// Check if admin
			isAdmin := false
			if sessInst, ok := r.Variables["$__session"].(*Instance); ok {
				if role, ok := sessInst.Fields["user_role"]; ok && role == "admin" {
					isAdmin = true
				}
			}
			if !isAdmin {
				return &Instance{
					Fields: map[string]interface{}{
						"_type": "REDIRECT",
						"url":   "/",
						"flash": map[string]interface{}{
							"error": "Acceso denegado.",
						},
					},
				}, nil
			}
		}
	}

	// Execute Handler
	// Execute Handler
	if handlerName, ok := handler.(string); ok {
		// Controller@Method
		parts := strings.Split(handlerName, "@")
		if len(parts) == 2 {
			controllerName := parts[0]
			methodName := parts[1]

			// Find Controller Class
			if classStmt, ok := r.Classes[controllerName]; ok {
				// Create Instance
				instance := &Instance{Class: classStmt, Fields: make(map[string]interface{})}

				// Find Method
				for _, stmt := range classStmt.Body.Statements {
					if m, ok := stmt.(*parser.MethodStatement); ok {
						if m.Name.Value == methodName {
							// Extract parameters if dynamic route
							args := []interface{}{}
							if strings.Contains(path, "/") { // Simple check, better logic needed for exact vs dynamic
								// Re-match to extract args if we haven't already
								// Optimization: We should have saved the matches/params earlier.
								// For now, let's re-match if it looks dynamic.
								for routePath := range r.Routes[method] {
									if strings.Contains(routePath, "{") {
										regexPath := "^" + regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`).ReplaceAllString(routePath, "([^/]+)") + "$"
										re := regexp.MustCompile(regexPath)
										matches := re.FindStringSubmatch(path)
										if len(matches) > 1 {
											// matches[1:] are the params
											for _, param := range matches[1:] {
												args = append(args, param)
											}
											break
										}
									}
								}
							}

							return r.CallMethodEvaluated(m, instance, args), nil
						}
					}
				}
				return nil, fmt.Errorf("method %s not found in controller %s", methodName, controllerName)
			}
			return nil, fmt.Errorf("controller %s not found", controllerName)
		}
	}

	return nil, nil
}
