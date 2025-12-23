package core

import "github.com/jossecurity/joss/pkg/parser"

// executeResponseMethod handles Response methods
func (r *Runtime) executeResponseMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "redirect":
		if len(args) > 0 {
			url := args[0].(string)
			return r.createWebResponse("REDIRECT", url, nil, 302)
		}
	case "back":
		referer := "/"
		if reqVal, ok := r.Variables["$__request"]; ok {
			if reqInstance, ok := reqVal.(*Instance); ok {
				if ref, ok := reqInstance.Fields["_referer"]; ok && ref != "" {
					referer = ref.(string)
				}
			}
		}
		return r.createWebResponse("REDIRECT", referer, nil, 302)

	case "json":
		if len(args) > 0 {
			statusCode := 200
			if len(args) > 1 {
				statusCode = toInt(args[1])
			}
			return r.createWebResponse("JSON", "", args[0], statusCode)
		}

	case "raw":
		if len(args) > 0 {
			content := args[0]
			statusCode := 200
			if len(args) > 1 {
				statusCode = toInt(args[1])
			}
			res := r.createWebResponse("RAW", "", content, statusCode)

			// Content Type
			if len(args) > 2 {
				if ct, ok := args[2].(string); ok {
					res.Fields["content_type"] = ct
				}
			}
			// Headers
			if len(args) > 3 {
				if headers, ok := args[3].(map[string]interface{}); ok {
					if resHeaders, ok := res.Fields["headers"].(map[string]interface{}); ok {
						for k, v := range headers {
							resHeaders[k] = v
						}
					}
				}
			}
			return res
		}
	}
	return nil
}

func (r *Runtime) createWebResponse(resType string, url string, data interface{}, status int) *Instance {
	// We need a class for this to support method calls
	if _, ok := r.Classes["WebResponse"]; !ok {
		r.registerClass(&parser.ClassStatement{
			Name: &parser.Identifier{Value: "WebResponse"},
			Body: &parser.BlockStatement{},
		})
	}

	instance := &Instance{
		Class: r.Classes["WebResponse"],
		Fields: map[string]interface{}{
			"_type":       resType,
			"url":         url,
			"data":        data,
			"status_code": status,
			"flash":       make(map[string]interface{}),
			"cookies":     make(map[string]interface{}),
			"headers":     make(map[string]interface{}),
		},
	}

	// Default Content Type for RAW
	if resType == "RAW" {
		instance.Fields["content_type"] = "text/plain"
	}

	return instance
}

// executeWebResponseMethod handles methods on the WebResponse instance (like ->withCookie())
func (r *Runtime) executeWebResponseMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "with":
		if len(args) >= 2 {
			key := args[0].(string)
			val := args[1]
			if flash, ok := instance.Fields["flash"].(map[string]interface{}); ok {
				flash[key] = val
			}
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

	case "withHeader":
		// ->withHeader(key, value)
		if len(args) >= 2 {
			key := args[0].(string)
			val := args[1]
			if headers, ok := instance.Fields["headers"].(map[string]interface{}); ok {
				headers[key] = val
			}
			return instance
		}

	case "status":
		// ->status(404)
		if len(args) >= 1 {
			instance.Fields["status_code"] = toInt(args[0])
			return instance
		}
	}
	return instance
}

func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 200
	}
}
