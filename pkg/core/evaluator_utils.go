package core

import (
	"fmt"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) checkType(val interface{}, typeName string) bool {
	if typeName == "" || typeName == "mixed" {
		return true
	}

	if val == nil {
		return false // Or allow null? For now, strict.
	}

	switch strings.ToLower(typeName) {
	case "int", "integer":
		switch v := val.(type) {
		case int, int32, int64:
			return true
		case float64:
			return v == float64(int64(v))
		}
		return false
	case "float", "double":
		switch val.(type) {
		case float64, float32, int, int64:
			return true
		}
		return false
	case "string":
		_, ok := val.(string)
		return ok
	case "bool", "boolean":
		_, ok := val.(bool)
		return ok
	case "array":
		_, ok := val.([]interface{})
		return ok
	case "map":
		_, ok := val.(map[string]interface{})
		return ok
	case "channel":
		_, ok := val.(*Channel)
		return ok
	case "object":
		_, ok := val.(*Instance)
		return ok
	default:
		// Check for specific class instance
		if inst, ok := val.(*Instance); ok {
			curr := inst.Class
			for curr != nil {
				if curr.Name.Value == typeName {
					return true
				}
				if curr.SuperClass != nil {
					if super, ok := r.Classes[curr.SuperClass.Value]; ok {
						curr = super
					} else {
						break
					}
				} else {
					break
				}
			}
		}
	}
	return false
}

func (r *Runtime) checkExistence(exp parser.Expression) bool {
	switch e := exp.(type) {
	case *parser.Identifier:
		_, ok := r.Variables[e.Value]
		return ok
	case *parser.IndexExpression:
		left := r.evaluateExpression(e.Left)
		if list, ok := left.([]interface{}); ok {
			index := r.evaluateExpression(e.Index)
			if idx, ok := index.(int64); ok {
				return idx >= 0 && idx < int64(len(list))
			}
		}
		return false
	case *parser.MemberExpression:
		left := r.evaluateExpression(e.Left)
		if instance, ok := left.(*Instance); ok {
			_, ok := instance.Fields[e.Property.Value]
			return ok
		}
		return false
	}
	return false
}

func isFalsy(val interface{}) bool {
	if val == nil {
		return true
	}
	if _, ok := val.(*Instance); ok {
		return false // Instances are always Truthy
	}
	if b, ok := val.(bool); ok {
		return !b
	}
	if s, ok := val.(string); ok {
		return s == "" || s == "0"
	}
	if i, ok := val.(int64); ok {
		return i == 0
	}
	if list, ok := val.([]interface{}); ok {
		return len(list) == 0
	}
	return false
}

func isTruthy(val interface{}) bool {
	return !isFalsy(val)
}

// coerceToTypedValue attempts to cast val to the declared type when val is a string.
// This allows Console::input() (which returns string) to work with int/float declarations.
func (r *Runtime) coerceToTypedValue(val interface{}, typeName string) interface{} {
	if val == nil {
		return val
	}
	str, isString := val.(string)
	if !isString {
		return val // Already a non-string, no coercion needed
	}
	switch strings.ToLower(typeName) {
	case "int", "integer":
		// Try to parse as int
		var i int64
		_, err := fmt.Sscanf(strings.TrimSpace(str), "%d", &i)
		if err == nil {
			return i
		}
		// Try float then truncate
		var f float64
		_, err = fmt.Sscanf(strings.TrimSpace(str), "%f", &f)
		if err == nil {
			return int64(f)
		}
	case "float", "double":
		var f float64
		_, err := fmt.Sscanf(strings.TrimSpace(str), "%f", &f)
		if err == nil {
			return f
		}
	case "bool", "boolean":
		s := strings.ToLower(strings.TrimSpace(str))
		if s == "true" || s == "1" || s == "yes" {
			return true
		}
		if s == "false" || s == "0" || s == "no" || s == "" {
			return false
		}
	}
	return val // Return original if no coercion possible
}

// getZeroValue returns the zero/default value for a given type name.
// Used when a variable is declared without an initializer (e.g., int $x).
func (r *Runtime) getZeroValue(typeName string) interface{} {
	switch strings.ToLower(typeName) {
	case "int", "integer":
		return int64(0)
	case "float", "double":
		return float64(0.0)
	case "string":
		return ""
	case "bool", "boolean":
		return false
	case "array":
		return []interface{}{}
	case "map":
		return map[string]interface{}{}
	default:
		return nil
	}
}
