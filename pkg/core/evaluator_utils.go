package core

import (
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

func (r *Runtime) checkType(val interface{}, typeName string) bool {
	destination := typesystem.Parse(typeName)
	source := runtimeTypeOf(val)
	if typesystem.Assignable(destination, source) {
		return true
	}
	if destination.Kind != typesystem.Class {
		return false
	}
	inst, ok := val.(*Instance)
	if !ok || inst == nil {
		return false
	}
	for class := inst.Class; class != nil; {
		if class.Name != nil && class.Name.Value == destination.Name {
			return true
		}
		if class.SuperClass == nil {
			break
		}
		class = r.Classes[class.SuperClass.Value]
	}
	return false
}

func runtimeTypeOf(value interface{}) typesystem.Type {
	switch typed := value.(type) {
	case nil:
		return typesystem.Type{Kind: typesystem.Null}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return typesystem.Type{Kind: typesystem.Int}
	case float32, float64:
		return typesystem.Type{Kind: typesystem.Float}
	case string:
		return typesystem.Type{Kind: typesystem.String}
	case bool:
		return typesystem.Type{Kind: typesystem.Bool}
	case []interface{}:
		return typesystem.Type{Kind: typesystem.Array}
	case map[string]interface{}:
		return typesystem.Type{Kind: typesystem.Map}
	case *Channel:
		return typesystem.Type{Kind: typesystem.Channel}
	case *Instance:
		if typed != nil && typed.Class != nil && typed.Class.Name != nil {
			return typesystem.Type{Kind: typesystem.Class, Name: typed.Class.Name.Value}
		}
		return typesystem.Type{Kind: typesystem.Object}
	default:
		return typesystem.Type{Kind: typesystem.Unknown}
	}
}

func runtimeTypeName(value interface{}) string {
	valueType := runtimeTypeOf(value)
	if !valueType.IsKnown() {
		return ""
	}
	return valueType.String()
}

func (r *Runtime) checkExistence(exp parser.Expression) bool {
	switch e := exp.(type) {
	case *parser.Identifier:
		_, ok := r.Variables[e.Value]
		return ok
	case *parser.IndexExpression:
		left := r.safeEvaluate(e.Left)
		if left == nil {
			return false
		}
		if list, ok := left.([]interface{}); ok {
			index := r.safeEvaluate(e.Index)
			if idx, ok := index.(int64); ok {
				return idx >= 0 && idx < int64(len(list))
			}
		}
		return false
	case *parser.MemberExpression:
		left := r.safeEvaluate(e.Left)
		if instance, ok := left.(*Instance); ok {
			_, ok := instance.Fields[e.Property.Value]
			return ok
		}
		return false
	}
	return false
}

// safeEvaluate evaluates an expression and returns nil if it panics.
// Used only for existence checks (isset/empty) where undefined is expected.
func (r *Runtime) safeEvaluate(exp parser.Expression) (result interface{}) {
	defer func() {
		if rec := recover(); rec != nil {
			result = nil
		}
	}()
	return r.evaluateExpression(exp)
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
	if coerced, ok := typesystem.CoerceString(typesystem.Parse(typeName), str); ok {
		return coerced
	}
	return val // Return original if no coercion possible
}

// getZeroValue returns the zero/default value for a given type name.
// Used when a variable is declared without an initializer (e.g., int $x).
func (r *Runtime) getZeroValue(typeName string) interface{} {
	switch typesystem.Parse(typeName).Kind {
	case typesystem.Int:
		return int64(0)
	case typesystem.Float:
		return float64(0.0)
	case typesystem.String:
		return ""
	case typesystem.Bool:
		return false
	case typesystem.Array:
		return []interface{}{}
	case typesystem.Map:
		return map[string]interface{}{}
	default:
		return nil
	}
}
