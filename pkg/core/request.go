package core

import "fmt"

// executeRequestMethod handles Request methods
func (r *Runtime) executeRequestMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {

	case "file":
		if len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return nil
			}

			// Access $__request variable
			if reqVal, ok := r.Variables["$__request"]; ok {
				if reqInstance, ok := reqVal.(*Instance); ok {
					// Check _files map
					if filesVal, ok := reqInstance.Fields["_files"]; ok {
						if filesMap, ok := filesVal.(map[string]interface{}); ok {
							if file, ok := filesMap[key]; ok {
								return file // Returns the map {name, content, ...}
							}
						}
					}
				}
			}
			return nil
		}

	case "input", "post":
		if len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return nil
			}

			// Access $__request variable injected by Dispatch
			if reqVal, ok := r.Variables["$__request"]; ok {
				if reqInstance, ok := reqVal.(*Instance); ok {
					if val, ok := reqInstance.Fields[key]; ok {
						return val
					}
				}
			}
			return nil
		}
	case "all":
		if reqVal, ok := r.Variables["$__request"]; ok {
			if reqInstance, ok := reqVal.(*Instance); ok {
				// Filter out internal fields
				result := make(map[string]interface{})
				for k, v := range reqInstance.Fields {
					// exclude internal fields starting with _
					if k != "_headers" && k != "_host" && k != "_scheme" && k != "_files" {
						result[k] = v
					}
				}
				return result
			}
		}
		return make(map[string]interface{})

	case "except":
		if len(args) > 0 {
			excludeMap := make(map[string]bool)
			if list, ok := args[0].([]interface{}); ok {
				for _, item := range list {
					excludeMap[fmt.Sprintf("%v", item)] = true
				}
			}

			if reqVal, ok := r.Variables["$__request"]; ok {
				if reqInstance, ok := reqVal.(*Instance); ok {
					result := make(map[string]interface{})
					for k, v := range reqInstance.Fields {
						// exclude internal fields
						if !excludeMap[k] && k != "_headers" && k != "_host" && k != "_scheme" && k != "_files" {
							result[k] = v
						}
					}
					return result
				}
			}
		}
		return make(map[string]interface{})

	case "root":
		// Return scheme://host
		if reqVal, ok := r.Variables["$__request"]; ok {
			if reqInstance, ok := reqVal.(*Instance); ok {
				scheme := "http"
				if s, ok := reqInstance.Fields["_scheme"].(string); ok {
					scheme = s
				}
				host := "localhost"
				if h, ok := reqInstance.Fields["_host"].(string); ok {
					host = h
				}
				return scheme + "://" + host
			}
		}
		return ""
	}
	return nil
}
