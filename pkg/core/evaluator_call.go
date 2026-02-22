package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	r.Variables["this"] = instance

	// Bind arguments
	for i, param := range method.Parameters {
		if i < len(args) {
			val := r.evaluateExpression(args[i])
			r.Variables[param.Value] = val
		}
	}

	defer func() {
		if prevThis != nil {
			r.Variables["this"] = prevThis
		} else {
			delete(r.Variables, "this")
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
	r.Variables["this"] = instance

	// Bind arguments
	for i, param := range method.Parameters {
		if i < len(args) {
			r.Variables[param.Value] = args[i]
		}
	}

	defer func() {
		if prevThis != nil {
			r.Variables["this"] = prevThis
		} else {
			delete(r.Variables, "this")
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

	// 2. Check Identifier for Builtins
	if ident, ok := call.Function.(*parser.Identifier); ok {
		if res, ok := r.callBuiltin(ident.Value, args); ok {
			return res
		}
	}

	// 3. Evaluate Function
	fn := r.evaluateExpression(call.Function)
	return r.applyFunction(fn, args)
}

func (r *Runtime) applyFunction(fn interface{}, args []interface{}) interface{} {
	if bound, ok := fn.(*BoundMethod); ok {
		if bound.Instance == nil && bound.StaticClass != "" {
			// Static Call
			// Evaluate arguments
			evalArgs := []interface{}{}
			for _, arg := range args {
				evalArgs = append(evalArgs, arg) // args are already evaluated in executeCall
			}
			// We need to call executeNativeMethod with a way to identify the class.
			// executeNativeMethod expects *Instance.
			// Let's create a temporary instance for the static call.
			// Or better, update executeNativeMethod to accept className string.
			// But that requires changing signature in native.go and all calls.
			// Easier: Create a dummy instance.
			dummyInstance := &Instance{
				Class: &parser.ClassStatement{
					Name: &parser.Identifier{Value: bound.StaticClass},
				},
			}
			return r.executeNativeMethod(dummyInstance, bound.Method.Name.Value, evalArgs)
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
	case "env":
		if len(args) == 1 {
			if key, ok := args[0].(string); ok {
				if val, exists := r.Env[key]; exists {
					return val, true
				}
				return "", true
			}
		}
		return nil, true
	case "len", "count":
		if len(args) == 1 {
			if list, ok := args[0].([]interface{}); ok {
				return int64(len(list)), true
			}
			if str, ok := args[0].(string); ok {
				return int64(len(str)), true
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
			go func() {
				defer func() {
					if p := recover(); p != nil {
						if rp, ok := p.(*ReturnPanic); ok {
							future.result = rp.Value
						} else {
							future.err = fmt.Errorf("%v", p)
						}
					}
					close(future.done)
				}()

				if fn, ok := argVal.(*parser.FunctionLiteral); ok {
					newR := r.Fork() // Thread-safety clone
					future.result = newR.executeBlock(fn.Body)
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
			fmt.Println("Error: await expects a Future")
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
	case "keys":
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
	case "values":
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
		if len(args) == 1 {
			if url, ok := args[0].(string); ok {
				return r.createWebResponse("REDIRECT", url, nil, 302), true
			}
		}
		return nil, true
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
	}
	return nil, false
}
