package core

// executeRequestMethod handles Request methods
func (r *Runtime) executeRequestMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "input":
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
				return reqInstance.Fields
			}
		}
		return make(map[string]interface{})
	}
	return nil
}
