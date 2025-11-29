package core

import (
	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) checkType(val interface{}, typeName string) bool {
	if val == nil {
		return true
	} // Allow nil?
	switch typeName {
	case "int":
		_, ok := val.(int64)
		return ok
	case "float":
		_, ok := val.(float64)
		return ok
	case "string":
		_, ok := val.(string)
		return ok
	case "bool":
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
	}
	return true // Unknown types (classes) not strictly checked yet
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
