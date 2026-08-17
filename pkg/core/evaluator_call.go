package core

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jossecurity/joss/pkg/i18n"
	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) CallMethod(method *parser.MethodStatement, instance *Instance, args []parser.Expression) (res interface{}) {
	// Native Method Support
	if method.Body == nil {
		evalArgs := []interface{}{}
		for _, arg := range args {
			evalArgs = append(evalArgs, r.evaluateExpression(arg))
		}

		// Check for Static Class Call
		if instance == nil {
			return nil
		}

		return r.executeNativeMethod(instance, method.Name.Value, evalArgs)
	}

	// Save previous "this" if exists (for nested calls)
	prevThis := r.Variables["this"]
	_, previousThisExists := r.Variables["this"]
	if instance != nil {
		r.Variables["this"] = instance
	}
	previousCaptureEnvironment := r.captureEnvironment
	r.captureEnvironment = nil

	// Bind arguments
	previousParams := make(map[string]interface{}, len(method.Parameters))
	previousParamExists := make(map[string]bool, len(method.Parameters))
	for i, param := range method.Parameters {
		previousParams[param.Name.Value], previousParamExists[param.Name.Value] = r.Variables[param.Name.Value]
		if i < len(args) {
			val := r.evaluateExpression(args[i])
			if param.Type.Literal != "" {
				if !r.checkType(val, param.Type.Literal) {
					panic(fmt.Sprintf("Type Error: El argumento %d (%s) debe ser de tipo %s, se recibió %T", i+1, param.Name.Value, param.Type.Literal, val))
				}
			}
			r.Variables[param.Name.Value] = val
		} else {
			r.Variables[param.Name.Value] = nil
		}
	}

	defer func() {
		r.captureEnvironment = previousCaptureEnvironment
		if instance != nil {
			if previousThisExists {
				r.Variables["this"] = prevThis
			} else {
				delete(r.Variables, "this")
			}
		}
		for _, param := range method.Parameters {
			if previousParamExists[param.Name.Value] {
				r.Variables[param.Name.Value] = previousParams[param.Name.Value]
			} else {
				delete(r.Variables, param.Name.Value)
			}
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			if rp, ok := p.(*ReturnPanic); ok {
				res = rp.Value
			} else {
				panic(p)
			}
		}
	}()

	return r.executeBlock(method.Body)
}

func (r *Runtime) CallMethodEvaluated(method *parser.MethodStatement, instance *Instance, args []interface{}) (res interface{}) {
	// Native Method Support
	if method.Body == nil {
		return r.executeNativeMethod(instance, method.Name.Value, args)
	}

	// Save previous "this" if exists (for nested calls)
	prevThis := r.Variables["this"]
	_, previousThisExists := r.Variables["this"]
	if instance != nil {
		r.Variables["this"] = instance
	}
	previousCaptureEnvironment := r.captureEnvironment
	r.captureEnvironment = nil

	// Bind arguments
	previousParams := make(map[string]interface{}, len(method.Parameters))
	previousParamExists := make(map[string]bool, len(method.Parameters))
	for i, param := range method.Parameters {
		previousParams[param.Name.Value], previousParamExists[param.Name.Value] = r.Variables[param.Name.Value]
		if i < len(args) {
			val := args[i]
			if param.Type.Literal != "" {
				if !r.checkType(val, param.Type.Literal) {
					panic(fmt.Sprintf("Type Error: El argumento %d (%s) debe ser de tipo %s, se recibió %T", i+1, param.Name.Value, param.Type.Literal, val))
				}
			}
			r.Variables[param.Name.Value] = val
		} else {
			r.Variables[param.Name.Value] = nil
		}
	}

	defer func() {
		r.captureEnvironment = previousCaptureEnvironment
		if instance != nil {
			if previousThisExists {
				r.Variables["this"] = prevThis
			} else {
				delete(r.Variables, "this")
			}
		}
		for _, param := range method.Parameters {
			if previousParamExists[param.Name.Value] {
				r.Variables[param.Name.Value] = previousParams[param.Name.Value]
			} else {
				delete(r.Variables, param.Name.Value)
			}
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			if rp, ok := p.(*ReturnPanic); ok {
				res = rp.Value
			} else {
				panic(p)
			}
		}
	}()

	return r.executeBlock(method.Body)
}

func (r *Runtime) executeCall(call *parser.CallExpression) interface{} {
	// 1. Evaluate arguments first
	args := []interface{}{}
	for _, arg := range call.Arguments {
		args = append(args, r.evaluateExpression(arg))
	}

	// 2. Try Builtin
	if ident, ok := call.Function.(*parser.Identifier); ok {
		if res, ok := r.callBuiltin(ident.Value, args); ok {
			return res
		}
	}

	// 3. Evaluate Function
	fn := r.evaluateExpression(call.Function)
	if fn == nil {
		if ident, ok := call.Function.(*parser.Identifier); ok {
			if f, ok := r.Functions[ident.Value]; ok {
				fn = f
			}
		}
	}

	if fn == nil {
		if ident, ok := call.Function.(*parser.Identifier); ok {
			panic(fmt.Sprintf("Error: Función '%s' no encontrada", ident.Value))
		}
		return nil
	}

	return r.applyFunction(fn, args)
}

func (r *Runtime) applyFunction(fn interface{}, args []interface{}) interface{} {
	if callable, ok := fn.(*PluginCallable); ok {
		if r.PluginRegistry != nil {
			if callable.ClassName != "" {
				res, err := r.PluginRegistry.CallMethod(callable.PluginName, callable.ClassName, callable.Function, nil, args)
				if err != nil {
					panic(fmt.Sprintf("Error en metodo de plugin %s::%s.%s: %v", callable.PluginName, callable.ClassName, callable.Function, err))
				}
				return res
			}
			res, err := r.PluginRegistry.CallFunction(callable.PluginName, callable.Function, args)
			if err != nil {
				panic(fmt.Sprintf("Error en funcion de plugin %s::%s: %v", callable.PluginName, callable.Function, err))
			}
			return res
		}
	}

	if closure, ok := fn.(*CapturedFunction); ok {
		return r.callCapturedFunction(closure, args)
	}

	if bound, ok := fn.(*BoundMethod); ok {
		if bound.Instance == nil && bound.StaticClass != "" {
			// Static Call
			evalArgs := []interface{}{}
			for _, arg := range args {
				evalArgs = append(evalArgs, arg)
			}
			classStmt := r.Classes[bound.StaticClass]
			if classStmt == nil {
				// Fallback to synthetic if not registered (should not happen for native)
				classStmt = &parser.ClassStatement{
					Name: &parser.Identifier{Value: bound.StaticClass},
					Body: &parser.BlockStatement{Statements: []parser.Statement{}},
				}
			}
			dummyInstance := &Instance{
				Class:  classStmt,
				Fields: make(map[string]interface{}),
			}
			return r.executeNativeMethod(dummyInstance, bound.Method.Name.Value, evalArgs)
		}

		if bound.Instance != nil && bound.Instance.Fields != nil && bound.Instance.Class != nil {
			if pluginName, isPlug := bound.Instance.Fields["__plugin__"].(string); isPlug && pluginName != "" {
				if r.PluginRegistry != nil {
					res, err := r.PluginRegistry.CallMethod(pluginName, bound.Instance.Class.Name.Value, bound.Method.Name.Value, bound.Instance, args)
					if err == nil {
						return res
					}
				}
			}
		}

		return r.CallMethodEvaluated(bound.Method, bound.Instance, args)
	}

	if method, ok := fn.(*parser.MethodStatement); ok {
		return r.CallMethodEvaluated(method, nil, args)
	}

	if lit, ok := fn.(*parser.FunctionLiteral); ok {
		// Create a synthetic method for the function literal
		method := &parser.MethodStatement{
			Token:      lit.Token,
			Name:       &parser.Identifier{Value: "anonymous"},
			Parameters: lit.Parameters,
			Body:       lit.Body,
		}
		return r.CallMethodEvaluated(method, nil, args)
	}

	if fn == nil {
		return nil
	}

	fmt.Printf("Error: '%v' (tipo %T) no es una función invocable\n", fn, fn)
	return nil
}

// Public API for executing functions (from Server etc)
func (r *Runtime) CallFunction(fn interface{}, args []interface{}) interface{} {
	return r.applyFunction(fn, args)
}

func (r *Runtime) callBuiltin(name string, args []interface{}) (interface{}, bool) {
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
	case "env", "config":
		if len(args) > 0 {
			if key, ok := args[0].(string); ok {
				if val, exists := r.Env[key]; exists && val != "" {
					return val, true
				}
				if len(args) > 1 {
					return args[1], true
				}
				if val, exists := r.Env[key]; exists {
					return val, true
				}
				return "", true
			}
		}
		return "", true
	case "view":
		return r.executeViewMethod(nil, "render", args), true
	case "json":
		return r.executeResponseMethod(nil, "json", args), true
	case "back":
		return r.executeResponseMethod(nil, "back", nil), true
	case "response":
		return r.executeResponseMethod(nil, "raw", args), true
	case "request":
		if len(args) == 0 {
			return r.executeRequestMethod(nil, "all", nil), true
		}
		return r.executeRequestMethod(nil, "input", args), true
	case "session":
		if len(args) == 0 {
			if sessVal, ok := r.Variables["$__session"]; ok {
				return sessVal, true
			}
			return nil, true
		}
		return r.executeSessionMethod(nil, "get", args), true
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
		// fmt.Printf("[callBuiltin empty] arg=%#v (type %T) -> resEmpty=%v\n", args[0], args[0], resEmpty)
		return resEmpty, true

	case "is_string":
		if len(args) == 1 {
			_, ok := args[0].(string)
			return ok, true
		}
		return false, true

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
			// Fallback usando reflection para cualquier slice o array
			val := reflect.ValueOf(args[0])
			if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
				return int64(val.Len()), true
			}
		}
		return int64(0), true
	case "toon_encode":
		if len(args) == 1 {
			return ToonEncode(args[0]), true
		}
		return "", true
	case "toon_decode":
		if len(args) == 1 {
			if str, ok := args[0].(string); ok {
				return ToonDecode(str), true
			}
		}
		return nil, true
	case "toon_verify":
		if len(args) == 1 {
			if str, ok := args[0].(string); ok {
				return ToonVerify(str), true
			}
		}
		return false, true
	case "json_encode":
		if len(args) == 1 {
			return JsonEncode(args[0]), true
		}
		return "", true
	case "json_decode":
		if len(args) == 1 {
			if str, ok := args[0].(string); ok {
				return JsonDecode(str), true
			}
		}
		return nil, true
	case "json_verify":
		if len(args) == 1 {
			if str, ok := args[0].(string); ok {
				return JsonVerify(str), true
			}
		}
		return false, true
	case "async":
		if len(args) == 1 {
			future := &Future{
				done: make(chan bool),
			}
			argVal := args[0]
			newR := r.Fork() // Fork BEFORE starting the goroutine to avoid race
			go func() {
				defer func() {
					if p := recover(); p != nil {
						if rp, ok := p.(*ReturnPanic); ok {
							future.result = rp.Value
						} else {
							fmt.Printf("[ASYNC PANIC] %v\n", p)
							future.err = fmt.Errorf("%v", p)
						}
					}
					close(future.done)
				}()

				if fn, ok := argVal.(*parser.FunctionLiteral); ok {
					future.result = newR.executeBlock(fn.Body)
				} else if blk, ok := argVal.(*parser.BlockStatement); ok {
					future.result = newR.executeBlock(blk)
				} else {
					future.result = argVal
				}
			}()
			return future, true
		}
		return nil, true
	case "await":
		if len(args) == 1 {
			if future, ok := args[0].(*Future); ok {
				return future.Wait(), true
			}
		}
		return nil, true
	case "make_chan":
		size := 0
		if len(args) > 0 {
			if s, ok := args[0].(int64); ok {
				size = int(s)
			}
		}
		return &Channel{Ch: make(chan interface{}, size)}, true
	case "close":
		if len(args) == 1 {
			if ch, ok := args[0].(*Channel); ok {
				close(ch.Ch)
				return nil, true
			}
		}
		return nil, true
	case "send":
		if len(args) == 2 {
			if ch, ok := args[0].(*Channel); ok {
				ch.Ch <- args[1]
				return nil, true
			}
		}
		return nil, true
	case "recv":
		if len(args) == 1 {
			if ch, ok := args[0].(*Channel); ok {
				val, ok := <-ch.Ch
				if !ok {
					return nil, true
				}
				return val, true
			}
		}
		return nil, true
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
	case "redirect":
		return r.executeResponseMethod(nil, "redirect", args), true
	case "explode":
		if len(args) == 2 {
			sep, ok1 := args[0].(string)
			str, ok2 := args[1].(string)
			if ok1 && ok2 {
				parts := strings.Split(str, sep)
				// Convert to []interface{}
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
	case "file_exists":
		if len(args) == 1 {
			if path, ok := args[0].(string); ok {
				_, err := os.Stat(path)
				return err == nil, true
			}
		}
		return false, true
	case "file_get_contents":
		if len(args) == 1 {
			if path, ok := args[0].(string); ok {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil, true
				}
				return string(content), true
			}
		}
		return nil, true
	case "file_put_contents":
		if len(args) == 2 {
			path, ok1 := args[0].(string)
			content, ok2 := args[1].(string)
			if ok1 && ok2 {
				err := os.WriteFile(path, []byte(content), 0644)
				if err != nil {
					return false, true
				}
				return true, true
			}
		}
		return false, true
	case "hive_read_box":
		if len(args) < 1 {
			return nil, true
		}
		filePath := fmt.Sprintf("%v", args[0])
		entries, err := ReadHiveBox(filePath)
		if err != nil {
			fmt.Printf("[hive_read_box Error] %v\n", err)
			return nil, true
		}
		result := make([]interface{}, len(entries))
		for i, entry := range entries {
			result[i] = entry
		}
		return result, true
	case "append":
		if len(args) == 2 {
			if list, ok := args[0].([]interface{}); ok {
				// Create new list to avoid mutating original if passed by value (slices are ref though)
				// But interface{} slice logic in Go:
				newList := append(list, args[1])
				return newList, true
			}
		}
		return nil, true
	case "merge":
		// merge(list1, list2)
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
	case "run":
		// run "script.py", args...
		if len(args) > 0 {
			scriptPath, ok := args[0].(string)
			if !ok {
				return "", true
			}

			// Security Check
			allow, ok := r.Env["ALLOW_SYSTEM_RUN"]
			if !ok || (allow != "true" && allow != "1") {
				fmt.Println("[Security] Error: Ejecución de scripts bloqueada. Configure ALLOW_SYSTEM_RUN=true en su entorno.")
				return "", true
			}

			// Determine runner
			runner := ""
			if strings.HasSuffix(scriptPath, ".py") {
				runner = "python"
			} else if strings.HasSuffix(scriptPath, ".php") {
				runner = "php"
			} else {
				fmt.Println("[Error] Tipo de archivo no soportado para 'run'. Use .py o .php")
				return "", true
			}

			// Build args
			cmdArgs := []string{scriptPath}
			// Add extra args
			if len(args) > 1 {
				for _, arg := range args[1:] {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%v", arg))
				}
			}

			cmd := exec.Command(runner, cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[Run] Error ejecutando script: %v\n", err)
			}
			return string(output), true
		}
		return "", true

	// --- Date & Time Functions ---
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

	// --- String Functions ---
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

	// --- Array & Map Functions ---
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

	// --- File & Directory Functions ---
	case "unlink", "file_delete":
		if len(args) > 0 {
			path := fmt.Sprintf("%v", args[0])
			err := os.Remove(path)
			return err == nil, true
		}
		return false, true
	case "mkdir":
		if len(args) > 0 {
			path := fmt.Sprintf("%v", args[0])
			err := os.MkdirAll(path, 0755)
			return err == nil, true
		}
		return false, true
	case "is_dir":
		if len(args) > 0 {
			path := fmt.Sprintf("%v", args[0])
			fi, err := os.Stat(path)
			return err == nil && fi.IsDir(), true
		}
		return false, true
	case "is_file":
		if len(args) > 0 {
			path := fmt.Sprintf("%v", args[0])
			fi, err := os.Stat(path)
			return err == nil && !fi.IsDir(), true
		}
		return false, true
	}
	return nil, false
}
