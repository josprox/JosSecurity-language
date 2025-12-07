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
				// Filter out _headers
				result := make(map[string]interface{})
				for k, v := range reqInstance.Fields {
					if k != "_headers" {
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
						if !excludeMap[k] && k != "_headers" {
							result[k] = v
						}
					}
					return result
				}
			}
		}
		return make(map[string]interface{})
	}
	return nil
}
