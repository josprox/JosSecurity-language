package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (r *Runtime) callBuiltinArray(name string, args []interface{}) (interface{}, bool) {
	switch name {
	case "isset":
		if len(args) == 0 || args[0] == nil {
			return false, true
		}
		if str, ok := args[0].(string); ok && str == "" {
			return false, true
		}
		return true, true

	case "empty":
		resEmpty := false
		if len(args) == 0 || args[0] == nil {
			resEmpty = true
		} else if b, ok := args[0].(bool); ok {
			resEmpty = !b
		} else if str, ok := args[0].(string); ok {
			resEmpty = (str == "" || str == "0")
		} else if num, ok := args[0].(int); ok {
			resEmpty = (num == 0)
		} else if num, ok := args[0].(int64); ok {
			resEmpty = (num == 0)
		} else if num, ok := args[0].(float64); ok {
			resEmpty = (num == 0)
		} else if list, ok := args[0].([]interface{}); ok {
			resEmpty = (len(list) == 0)
		} else if m, ok := args[0].(map[string]interface{}); ok {
			resEmpty = (len(m) == 0)
		} else {
			val := reflect.ValueOf(args[0])
			if val.Kind() == reflect.Slice || val.Kind() == reflect.Array || val.Kind() == reflect.Map {
				resEmpty = (val.Len() == 0)
			}
		}
		return resEmpty, true

	case "is_string":
		if len(args) == 1 {
			_, ok := args[0].(string)
			return ok, true
		}
		return false, true

	case "is_numeric":
		if len(args) == 1 {
			if args[0] == nil {
				return false, true
			}
			switch v := args[0].(type) {
			case int, int32, int64, float64, float32:
				return true, true
			case string:
				vStr := strings.TrimSpace(v)
				if _, err := strconv.ParseFloat(vStr, 64); err == nil {
					return true, true
				}
				return false, true
			}
		}
		return false, true

	case "is_int", "is_integer":
		if len(args) == 1 {
			switch args[0].(type) {
			case int, int32, int64:
				return true, true
			}
		}
		return false, true

	case "is_float", "is_double":
		if len(args) == 1 {
			switch args[0].(type) {
			case float64, float32:
				return true, true
			}
		}
		return false, true

	case "intval":
		if len(args) == 0 || args[0] == nil {
			return int64(0), true
		}
		switch v := args[0].(type) {
		case int:
			return int64(v), true
		case int32:
			return int64(v), true
		case int64:
			return v, true
		case float64:
			return int64(v), true
		case float32:
			return int64(v), true
		case bool:
			if v {
				return int64(1), true
			}
			return int64(0), true
		case string:
			vStr := strings.TrimSpace(v)
			if n, err := strconv.ParseInt(vStr, 10, 64); err == nil {
				return n, true
			}
			if f, err := strconv.ParseFloat(vStr, 64); err == nil {
				return int64(f), true
			}
		}
		return int64(0), true

	case "floatval", "doubleval":
		if len(args) == 0 || args[0] == nil {
			return float64(0.0), true
		}
		switch v := args[0].(type) {
		case float64:
			return v, true
		case float32:
			return float64(v), true
		case int:
			return float64(v), true
		case int32:
			return float64(v), true
		case int64:
			return float64(v), true
		case bool:
			if v {
				return float64(1.0), true
			}
			return float64(0.0), true
		case string:
			vStr := strings.TrimSpace(v)
			if f, err := strconv.ParseFloat(vStr, 64); err == nil {
				return f, true
			}
		}
		return float64(0.0), true

	case "strval":
		if len(args) == 0 || args[0] == nil {
			return "", true
		}
		return fmt.Sprintf("%v", args[0]), true

	case "boolval":
		if len(args) == 0 || args[0] == nil {
			return false, true
		}
		return isTruthy(args[0]), true

	case "is_array":
		if len(args) == 1 {
			if _, ok := args[0].([]interface{}); ok {
				return true, true
			}
			val := reflect.ValueOf(args[0])
			return val.Kind() == reflect.Slice || val.Kind() == reflect.Array, true
		}
		return false, true

	case "is_null":
		if len(args) == 1 {
			return args[0] == nil, true
		}
		return true, true

	case "len", "count":
		if len(args) == 1 {
			if args[0] == nil {
				return int64(0), true
			}
			if list, ok := args[0].([]interface{}); ok {
				return int64(len(list)), true
			}
			if listMap, ok := args[0].([]map[string]interface{}); ok {
				return int64(len(listMap)), true
			}
			if m, ok := args[0].(map[string]interface{}); ok {
				return int64(len(m)), true
			}
			if str, ok := args[0].(string); ok {
				return int64(len(str)), true
			}
			val := reflect.ValueOf(args[0])
			if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
				return int64(val.Len()), true
			}
		}
		return int64(0), true

	case "keys", "array_keys":
		if len(args) == 1 {
			if m, ok := args[0].(map[string]interface{}); ok {
				keys := []interface{}{}
				for k := range m {
					keys = append(keys, k)
				}
				return keys, true
			}
		}
		return []interface{}{}, true

	case "values", "array_values":
		if len(args) == 1 {
			if m, ok := args[0].(map[string]interface{}); ok {
				vals := []interface{}{}
				for _, v := range m {
					vals = append(vals, v)
				}
				return vals, true
			}
		}
		return []interface{}{}, true

	case "explode":
		if len(args) == 2 {
			sep, ok1 := args[0].(string)
			str, ok2 := args[1].(string)
			if ok1 && ok2 {
				parts := strings.Split(str, sep)
				result := []interface{}{}
				for _, p := range parts {
					result = append(result, p)
				}
				return result, true
			}
		}
		return nil, true

	case "end":
		if len(args) == 1 {
			if list, ok := args[0].([]interface{}); ok {
				if len(list) > 0 {
					return list[len(list)-1], true
				}
				return nil, true
			}
		}
		return nil, true

	case "append":
		if len(args) == 2 {
			if list, ok := args[0].([]interface{}); ok {
				newList := append(list, args[1])
				return newList, true
			}
		}
		return nil, true

	case "merge":
		if len(args) == 2 {
			l1, ok1 := args[0].([]interface{})
			l2, ok2 := args[1].([]interface{})
			if ok1 && ok2 {
				newList := make([]interface{}, len(l1)+len(l2))
				copy(newList, l1)
				copy(newList[len(l1):], l2)
				return newList, true
			}
		}
		return nil, true

	case "in_array":
		if len(args) >= 2 {
			target := args[0]
			if list, ok := args[1].([]interface{}); ok {
				for _, item := range list {
					if reflect.DeepEqual(item, target) || fmt.Sprintf("%v", item) == fmt.Sprintf("%v", target) {
						return true, true
					}
				}
				return false, true
			}
			val := reflect.ValueOf(args[1])
			if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
				for i := 0; i < val.Len(); i++ {
					if reflect.DeepEqual(val.Index(i).Interface(), target) {
						return true, true
					}
				}
			}
		}
		return false, true

	case "array_key_exists":
		if len(args) >= 2 {
			k := fmt.Sprintf("%v", args[0])
			if m, ok := args[1].(map[string]interface{}); ok {
				_, exists := m[k]
				return exists, true
			}
		}
		return false, true

	case "array_merge":
		if len(args) >= 2 {
			if l1, ok1 := args[0].([]interface{}); ok1 {
				res := append([]interface{}{}, l1...)
				for _, next := range args[1:] {
					if lNext, okN := next.([]interface{}); okN {
						res = append(res, lNext...)
					}
				}
				return res, true
			}
			if m1, ok1 := args[0].(map[string]interface{}); ok1 {
				res := make(map[string]interface{}, len(m1))
				for k, v := range m1 {
					res[k] = v
				}
				for _, next := range args[1:] {
					if mNext, okN := next.(map[string]interface{}); okN {
						for k, v := range mNext {
							res[k] = v
						}
					}
				}
				return res, true
			}
		}
		return []interface{}{}, true

	case "array_push":
		if len(args) >= 2 {
			if list, ok := args[0].([]interface{}); ok {
				return append(list, args[1:]...), true
			}
		}
		return nil, true

	case "array_pop":
		if len(args) >= 1 {
			if list, ok := args[0].([]interface{}); ok && len(list) > 0 {
				return list[len(list)-1], true
			}
		}
		return nil, true

	case "array_shift":
		if len(args) >= 1 {
			if list, ok := args[0].([]interface{}); ok && len(list) > 0 {
				return list[0], true
			}
		}
		return nil, true

	case "array_slice":
		if len(args) >= 2 {
			if list, ok := args[0].([]interface{}); ok {
				offset := 0
				switch v := args[1].(type) {
				case int:
					offset = v
				case int64:
					offset = int(v)
				}
				l := len(list)
				if offset < 0 {
					offset = l + offset
					if offset < 0 {
						offset = 0
					}
				}
				if offset > l {
					return []interface{}{}, true
				}
				if len(args) >= 3 {
					length := 0
					switch v := args[2].(type) {
					case int:
						length = v
					case int64:
						length = int(v)
					}
					if length < 0 {
						end := l + length
						if end <= offset {
							return []interface{}{}, true
						}
						return list[offset:end], true
					}
					end := offset + length
					if end > l {
						end = l
					}
					return list[offset:end], true
				}
				return list[offset:], true
			}
		}
		return []interface{}{}, true
	}

	return nil, false
}
