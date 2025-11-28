package core

import (
	"fmt"
	"strings"
)

// executeRouterMethod handles Router class methods (get, post, api, match)
// This fixes the bug where Router methods were not implemented
func (r *Runtime) executeRouterMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Initialize Routes map if needed
	if r.Routes == nil {
		r.Routes = make(map[string]map[string]interface{})
	}

	// Helper to add route
	addRoute := func(method, path string, handler interface{}) {
		if r.Routes[method] == nil {
			r.Routes[method] = make(map[string]interface{})
		}

		// Store as RouteInfo map
		routeInfo := map[string]interface{}{
			"handler":    handler,
			"middleware": []string{},
		}

		// Add current middleware if any
		if r.CurrentMiddleware != nil && len(r.CurrentMiddleware) > 0 {
			mwCopy := make([]string, len(r.CurrentMiddleware))
			copy(mwCopy, r.CurrentMiddleware)
			routeInfo["middleware"] = mwCopy
		}

		r.Routes[method][path] = routeInfo
	}

	switch method {
	case "middleware":
		if len(args) >= 1 {
			if mw, ok := args[0].(string); ok {
				// Start middleware group
				if r.CurrentMiddleware == nil {
					r.CurrentMiddleware = []string{}
				}
				r.CurrentMiddleware = append(r.CurrentMiddleware, mw)
			}
		}
		return nil

	case "end":
		// End middleware group (pop last)
		if r.CurrentMiddleware != nil && len(r.CurrentMiddleware) > 0 {
			r.CurrentMiddleware = r.CurrentMiddleware[:len(r.CurrentMiddleware)-1]
		}
		return nil

	case "get":
		if len(args) >= 2 {
			path := args[0].(string)
			handler := args[1]
			addRoute("GET", path, handler)
			fmt.Printf("[DEBUG] executeRouterMethod called: get (%s)\n", path)
		}
		return nil

	case "post":
		if len(args) >= 2 {
			path := args[0].(string)
			handler := args[1]
			addRoute("POST", path, handler)
			fmt.Printf("[DEBUG] executeRouterMethod called: post (%s)\n", path)
		}
		return nil

	case "put":
		if len(args) >= 2 {
			path := args[0].(string)
			handler := args[1]
			addRoute("PUT", path, handler)
			fmt.Printf("[DEBUG] executeRouterMethod called: put (%s)\n", path)
		}
		return nil

	case "delete":
		if len(args) >= 2 {
			path := args[0].(string)
			handler := args[1]
			addRoute("DELETE", path, handler)
			fmt.Printf("[DEBUG] executeRouterMethod called: delete (%s)\n", path)
		}
		return nil

	case "match":
		// Router::match("GET|POST", "/path", "Controller@method1@method2")
		if len(args) >= 3 {
			methodsStr := args[0].(string)
			path := args[1].(string)
			handlerStr := args[2].(string)

			methods := strings.Split(methodsStr, "|")
			handlerParts := strings.Split(handlerStr, "@")

			// Case 1: Controller@method (Same for all)
			if len(handlerParts) == 2 {
				for _, m := range methods {
					addRoute(strings.ToUpper(strings.TrimSpace(m)), path, handlerStr)
				}
			} else if len(handlerParts) > 2 {
				// Case 2: Controller@method1@method2 (Map to methods)
				controller := handlerParts[0]
				methodHandlers := handlerParts[1:]

				for i, m := range methods {
					if i < len(methodHandlers) {
						fullHandler := controller + "@" + methodHandlers[i]
						addRoute(strings.ToUpper(strings.TrimSpace(m)), path, fullHandler)
					}
				}
			}
			fmt.Printf("[DEBUG] executeRouterMethod called: match (%s)\n", path)
		}
		return nil

	case "api":
		// API routes can be GET or POST, register for both
		if len(args) >= 2 {
			path := args[0].(string)
			handler := args[1]
			addRoute("GET", path, handler)
			addRoute("POST", path, handler)
			fmt.Printf("[DEBUG] executeRouterMethod called: api (%s)\n", path)
		}
		return nil
	}

	return nil
}
