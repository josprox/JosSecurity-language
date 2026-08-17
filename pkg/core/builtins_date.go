package core

import (
	"fmt"
	"strconv"
	"time"
)

func (r *Runtime) callBuiltinDate(name string, args []interface{}) (interface{}, bool) {
	switch name {
	case "time":
		return time.Now().Unix(), true

	case "microtime":
		if len(args) > 0 {
			if asFloat, ok := args[0].(bool); ok && asFloat {
				return float64(time.Now().UnixNano()) / 1e9, true
			}
		}
		now := time.Now()
		sec := now.Unix()
		usec := float64(now.Nanosecond()) / 1e9
		return fmt.Sprintf("%.8f %d", usec, sec), true

	case "date":
		format := "Y-m-d H:i:s"
		t := time.Now()
		if len(args) > 0 {
			if f, ok := args[0].(string); ok {
				format = f
			}
		}
		if len(args) > 1 && args[1] != nil {
			switch val := args[1].(type) {
			case int64:
				t = time.Unix(val, 0)
			case int:
				t = time.Unix(int64(val), 0)
			case float64:
				t = time.Unix(int64(val), 0)
			case string:
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					t = time.Unix(n, 0)
				} else if parsed, err := ParseHumanTime(val, time.Now()); err == nil {
					t = parsed
				}
			case time.Time:
				t = val
			}
		}
		return FormatDate(format, t), true

	case "strtotime":
		if len(args) == 0 || args[0] == nil {
			return nil, true
		}
		timeStr := fmt.Sprintf("%v", args[0])
		base := time.Now()
		if len(args) > 1 && args[1] != nil {
			switch bVal := args[1].(type) {
			case int64:
				base = time.Unix(bVal, 0)
			case int:
				base = time.Unix(int64(bVal), 0)
			case float64:
				base = time.Unix(int64(bVal), 0)
			}
		}
		parsed, err := ParseHumanTime(timeStr, base)
		if err != nil {
			return nil, true
		}
		return parsed.Unix(), true

	case "now":
		format := "Y-m-d H:i:s"
		if len(args) > 0 {
			if f, ok := args[0].(string); ok {
				format = f
			}
		}
		return FormatDate(format, time.Now()), true

	case "sleep":
		if len(args) > 0 {
			sec := 0
			switch v := args[0].(type) {
			case int:
				sec = v
			case int64:
				sec = int(v)
			case float64:
				time.Sleep(time.Duration(v * float64(time.Second)))
				return nil, true
			}
			if sec > 0 {
				time.Sleep(time.Duration(sec) * time.Second)
			}
		}
		return nil, true

	case "usleep":
		if len(args) > 0 {
			usec := int64(0)
			switch v := args[0].(type) {
			case int:
				usec = int64(v)
			case int64:
				usec = v
			case float64:
				usec = int64(v)
			}
			if usec > 0 {
				time.Sleep(time.Duration(usec) * time.Microsecond)
			}
		}
		return nil, true
	}

	return nil, false
}
