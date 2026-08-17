package core

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"strings"

	"github.com/jossecurity/joss/pkg/i18n"
)

func (r *Runtime) callBuiltinString(name string, args []interface{}) (interface{}, bool) {
	switch name {
	case "html_escape":
		if len(args) > 0 {
			if args[0] == nil {
				return "", true
			}
			return html.EscapeString(fmt.Sprintf("%v", args[0])), true
		}
		return "", true

	case "__":
		if len(args) > 0 {
			if key, ok := args[0].(string); ok {
				locale := r.GetLocale()
				return i18n.GlobalManager.Get(locale, key, nil), true
			}
		}
		return "", true

	case "csrf_field":
		tokenVal := ""
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if tok, ok := sessInst.Fields["csrf_token"]; ok {
					tokenVal = fmt.Sprintf("%v", tok)
				}
			}
		}
		return fmt.Sprintf(`<input type="hidden" name="_token" value="%s">`, tokenVal), true

	case "print", "echo":
		for _, arg := range args {
			fmt.Println(arg)
		}
		return nil, true

	case "printf":
		if len(args) > 0 {
			if fmtStr, ok := args[0].(string); ok {
				fmt.Printf(fmtStr, args[1:]...)
			}
		}
		return nil, true

	case "str_contains", "contains":
		if len(args) >= 2 {
			s1 := fmt.Sprintf("%v", args[0])
			s2 := fmt.Sprintf("%v", args[1])
			return strings.Contains(s1, s2), true
		}
		return false, true

	case "str_starts_with", "starts_with":
		if len(args) >= 2 {
			s1 := fmt.Sprintf("%v", args[0])
			s2 := fmt.Sprintf("%v", args[1])
			return strings.HasPrefix(s1, s2), true
		}
		return false, true

	case "str_ends_with", "ends_with":
		if len(args) >= 2 {
			s1 := fmt.Sprintf("%v", args[0])
			s2 := fmt.Sprintf("%v", args[1])
			return strings.HasSuffix(s1, s2), true
		}
		return false, true

	case "str_replace":
		if len(args) >= 3 {
			search := fmt.Sprintf("%v", args[0])
			replace := fmt.Sprintf("%v", args[1])
			subject := fmt.Sprintf("%v", args[2])
			return strings.ReplaceAll(subject, search, replace), true
		}
		return "", true

	case "strtolower", "to_lower":
		if len(args) > 0 {
			return strings.ToLower(fmt.Sprintf("%v", args[0])), true
		}
		return "", true

	case "strtoupper", "to_upper":
		if len(args) > 0 {
			return strings.ToUpper(fmt.Sprintf("%v", args[0])), true
		}
		return "", true

	case "trim":
		if len(args) > 0 {
			str := fmt.Sprintf("%v", args[0])
			if len(args) > 1 {
				cutset := fmt.Sprintf("%v", args[1])
				return strings.Trim(str, cutset), true
			}
			return strings.TrimSpace(str), true
		}
		return "", true

	case "ltrim":
		if len(args) > 0 {
			str := fmt.Sprintf("%v", args[0])
			if len(args) > 1 {
				cutset := fmt.Sprintf("%v", args[1])
				return strings.TrimLeft(str, cutset), true
			}
			return strings.TrimLeft(str, " \t\n\r\v\f"), true
		}
		return "", true

	case "rtrim":
		if len(args) > 0 {
			str := fmt.Sprintf("%v", args[0])
			if len(args) > 1 {
				cutset := fmt.Sprintf("%v", args[1])
				return strings.TrimRight(str, cutset), true
			}
			return strings.TrimRight(str, " \t\n\r\v\f"), true
		}
		return "", true

	case "substr":
		if len(args) >= 2 {
			str := fmt.Sprintf("%v", args[0])
			start := 0
			switch v := args[1].(type) {
			case int:
				start = v
			case int64:
				start = int(v)
			}
			runes := []rune(str)
			rLen := len(runes)
			if start < 0 {
				start = rLen + start
				if start < 0 {
					start = 0
				}
			}
			if start > rLen {
				return "", true
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
					end := rLen + length
					if end <= start {
						return "", true
					}
					return string(runes[start:end]), true
				}
				end := start + length
				if end > rLen {
					end = rLen
				}
				return string(runes[start:end]), true
			}
			return string(runes[start:]), true
		}
		return "", true

	case "strpos":
		if len(args) >= 2 {
			haystack := fmt.Sprintf("%v", args[0])
			needle := fmt.Sprintf("%v", args[1])
			idx := strings.Index(haystack, needle)
			if idx == -1 {
				return false, true
			}
			return int64(idx), true
		}
		return false, true

	case "implode", "join":
		if len(args) >= 2 {
			glue := fmt.Sprintf("%v", args[0])
			if list, ok := args[1].([]interface{}); ok {
				strs := make([]string, len(list))
				for i, item := range list {
					strs[i] = fmt.Sprintf("%v", item)
				}
				return strings.Join(strs, glue), true
			}
		}
		return "", true

	case "md5":
		if len(args) > 0 {
			h := md5.Sum([]byte(fmt.Sprintf("%v", args[0])))
			return hex.EncodeToString(h[:]), true
		}
		return "", true

	case "sha1":
		if len(args) > 0 {
			h := sha1.Sum([]byte(fmt.Sprintf("%v", args[0])))
			return hex.EncodeToString(h[:]), true
		}
		return "", true

	case "sha256":
		if len(args) > 0 {
			h := sha256.Sum256([]byte(fmt.Sprintf("%v", args[0])))
			return hex.EncodeToString(h[:]), true
		}
		return "", true

	case "base64_encode":
		if len(args) > 0 {
			return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", args[0]))), true
		}
		return "", true

	case "base64_decode":
		if len(args) > 0 {
			b, err := base64.StdEncoding.DecodeString(fmt.Sprintf("%v", args[0]))
			if err != nil {
				return false, true
			}
			return string(b), true
		}
		return false, true
	}

	return nil, false
}
