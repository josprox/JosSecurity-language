package core

import "github.com/jossecurity/joss/pkg/parser"

// executeResponseMethod handles Response methods
func (r *Runtime) executeResponseMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "redirect":
		if len(args) > 0 {
			url := args[0].(string)
			return r.createRedirectResponse(url)
		}
	case "back":
		// Try to get referer from request
		referer := "/"
		if reqVal, ok := r.Variables["$__request"]; ok {
			if reqInstance, ok := reqVal.(*Instance); ok {
				if ref, ok := reqInstance.Fields["_referer"]; ok && ref != "" {
					referer = ref.(string)
				}
			}
		}

		return r.createRedirectResponse(referer)

	case "json":
		if len(args) > 0 {
			res := map[string]interface{}{
				"_type": "JSON",
				"data":  args[0],
			}
			if len(args) > 1 {
				res["status_code"] = args[1]
			}
			return res
		}
	case "raw":
		if len(args) > 0 {
			res := map[string]interface{}{
				"_type": "RAW",
				"data":  args[0],
			}
			res["status_code"] = 200
			res["content_type"] = "text/plain"

			if len(args) > 1 {
				res["status_code"] = args[1]
			}
			if len(args) > 2 {
				res["content_type"] = args[2]
			}
			if len(args) > 3 {
				res["headers"] = args[3]
			}
			return res
		}

	}
	return nil
}

func (r *Runtime) createRedirectResponse(url string) *Instance {
	// We need a class for this to support method calls
	if _, ok := r.Classes["RedirectResponse"]; !ok {
		// Register it lazily if not exists (though better in native.go)
		r.registerClass(&parser.ClassStatement{
			Name: &parser.Identifier{Value: "RedirectResponse"},
			Body: &parser.BlockStatement{},
		})
	}

	instance := &Instance{
		Class: r.Classes["RedirectResponse"],
		Fields: map[string]interface{}{
			"_type":   "REDIRECT",
			"url":     url,
			"flash":   make(map[string]interface{}),
			"cookies": make(map[string]interface{}),
		},
	}
	return instance
}

// executeRedirectResponseMethod handles methods on the RedirectResponse instance (like ->with())
func (r *Runtime) executeRedirectResponseMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "with":
		if len(args) >= 2 {
			key := args[0].(string)
			val := args[1]

			if flash, ok := instance.Fields["flash"].(map[string]interface{}); ok {
				flash[key] = val
			}

			// Return the instance itself to allow chaining
			return instance
		}

	case "withCookie":
		// ->withCookie(name, value)
		if len(args) >= 2 {
			key := args[0].(string)
			val := args[1]

			if cookies, ok := instance.Fields["cookies"].(map[string]interface{}); ok {
				cookies[key] = val
			}
			return instance
		}
	}
	return instance
}
