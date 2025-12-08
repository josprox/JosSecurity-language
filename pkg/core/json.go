package core

import (
	"encoding/json"
)

// executeJSONMethod handles JSON methods (parse, stringify)
func (r *Runtime) executeJSONMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "parse", "decode":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok {
				var result interface{}
				// Use UseNumber to preserve number precision if needed, but standard Unmarshal is usually fine for basic types
				if err := json.Unmarshal([]byte(str), &result); err == nil {
					return result
				}
			}
		}
		return nil

	case "stringify", "encode":
		if len(args) > 0 {
			if bytes, err := json.Marshal(args[0]); err == nil {
				return string(bytes)
			}
		}
		return ""
	}
	return nil
}
