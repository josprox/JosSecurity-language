package core

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
		fmt.Printf("[DEBUG] DISPATCH FAIL: Route not found: %s %s\n", method, path)
		fmt.Printf("[DEBUG] Available Routes for %s:\n", method)
		if r.Routes[method] != nil {
			for k := range r.Routes[method] {
				fmt.Printf("\t'%s'\n", k)
			}
		} else {
			fmt.Println("\t(None)")
		}
		return nil, fmt.Errorf("route not found: %s %s", method, path)
	}

	// Debug Handler Type
	fmt.Printf("[DEBUG] DISPATCH SUCCESS: Found handler for %s %s\n", method, path)

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
		case "auth_api":
			// Check Authorization header
			authHeader := ""
			// We need to access headers securely. ReqData should have it.
			// Assuming reqData["header"] or reqData["headers"]
			if reqInst, ok := r.Variables["$__request"].(*Instance); ok {
				// Try to find headers
				if h, ok := reqInst.Fields["_headers"].(map[string]interface{}); ok {
					if val, k := h["Authorization"].(string); k {
						authHeader = val
					}
				}
				// Fallback or variation
				if authHeader == "" {
					if h, ok := reqData["Authorization"].(string); ok {
						authHeader = h
					}
				}
			}

			// Simple Bearer check
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				return &Instance{Fields: map[string]interface{}{
					"_type":  "JSON",
					"status": 401,
					"data":   map[string]interface{}{"error": "Unauthorized: Missing Token"},
				}}, nil
			}

			// Verify JWT
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				secret := os.Getenv("JWT_SECRET")
				if secret == "" {
					secret = "joss_default_secret_change_in_production"
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return &Instance{Fields: map[string]interface{}{
					"_type":  "JSON",
					"status": 401,
					"data":   map[string]interface{}{"error": "Unauthorized: Invalid Token"},
				}}, nil
			}

			// Success: Inject user info into request?
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Inject into $__user or similar if needed. For now, pass.
				// Maybe populate $__session with user info for this request context?
				if sessInst, ok := r.Variables["$__session"].(*Instance); ok {
					if uid, ok := claims["user_id"].(float64); ok {
						sessInst.Fields["user_id"] = int(uid)
					} else {
						sessInst.Fields["user_id"] = claims["user_id"]
					}
					sessInst.Fields["user_email"] = claims["email"]
					sessInst.Fields["user_name"] = claims["name"]
				}
			}

		default:
			// Custom Middleware
			if r.CustomMiddlewares != nil {
				if handler, ok := r.CustomMiddlewares[mw]; ok {
					// Execute closure
					res := r.applyFunction(handler, []interface{}{})
					// If returns a Result/Response, stop and return it
					if inst, ok := res.(*Instance); ok {
						// Assuming standard response structure or check existence
						if _, hasType := inst.Fields["_type"]; hasType {
							return inst, nil
						}
					}
				}
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
							if strings.Contains(path, "/") {
								for routePath := range r.Routes[method] {
									if strings.Contains(routePath, "{") {
										regexPath := "^" + regexp.MustCompile(`\{[a-zA-Z0-9_]+\}`).ReplaceAllString(routePath, "([^/]+)") + "$"
										re := regexp.MustCompile(regexPath)
										matches := re.FindStringSubmatch(path)
										if len(matches) > 1 {
											for _, param := range matches[1:] {
												args = append(args, param)
											}
											break
										}
									}
								}
							}
							fmt.Printf("[DEBUG] Executing method %s@%s\n", controllerName, methodName)
							return r.CallMethodEvaluated(m, instance, args), nil
						}
					}
				}
				fmt.Printf("[DEBUG] Method %s not found in controller %s\n", methodName, controllerName)
				return nil, fmt.Errorf("method %s not found in controller %s", methodName, controllerName)
			}
			fmt.Printf("[DEBUG] Controller %s not found. Available classes: %d\n", controllerName, len(r.Classes))
			for k := range r.Classes {
				fmt.Printf("\tClass: %s\n", k)
			}
			return nil, fmt.Errorf("controller %s not found", controllerName)
		}
	}

	return nil, nil
}
