package core

import (
	"fmt"
	"reflect"
)

func (r *Runtime) isNativeClass(name string) bool {
	if r == nil {
		return false
	}
	if _, ok := r.NativeHandlers[name]; ok {
		return true
	}
	switch name {
	case "Schema", "Blueprint", "Migration":
		return true
	}
	return false
}

func strictCompare(a, b interface{}) bool {
	a = normalizeNumber(a)
	b = normalizeNumber(b)
	if a == nil || b == nil {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}

func normalizeNumber(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case int16:
		return int64(val)
	case int8:
		return int64(val)
	case uint:
		return int64(val)
	case uint64:
		return int64(val)
	case uint32:
		return int64(val)
	case uint16:
		return int64(val)
	case uint8:
		return int64(val)
	case float32:
		return float64(val)
	}
	return v
}

func spaceshipCompare(left, right interface{}) int64 {
	a := normalizeNumber(left)
	b := normalizeNumber(right)
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	switch aVal := a.(type) {
	case int64:
		if bVal, ok := b.(int64); ok {
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		}
		if bVal, ok := b.(float64); ok {
			fVal := float64(aVal)
			if fVal < bVal {
				return -1
			}
			if fVal > bVal {
				return 1
			}
			return 0
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		}
		if bVal, ok := b.(int64); ok {
			fVal := float64(bVal)
			if aVal < fVal {
				return -1
			}
			if aVal > fVal {
				return 1
			}
			return 0
		}
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		}
	}
	aStr := fmt.Sprintf("%v", left)
	bStr := fmt.Sprintf("%v", right)
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}

func compareLessThan(a, b interface{}) bool {
	return spaceshipCompare(a, b) < 0
}

func compareGreaterThan(a, b interface{}) bool {
	return spaceshipCompare(a, b) > 0
}

func toInt64Safe(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case float64:
		return int64(val), true
	case float32:
		return int64(val), true
	case string:
		var n int64
		if _, err := fmt.Sscanf(val, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}
