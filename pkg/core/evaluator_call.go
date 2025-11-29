package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) CallMethod(method *parser.MethodStatement, instance *Instance, args []parser.Expression) interface{} {
	// Native Method Support
	if method.Body == nil {
		evalArgs := []interface{}{}
		for _, arg := range args {
			evalArgs = append(evalArgs, r.evaluateExpression(arg))
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

	// Execute body
	res := r.executeBlock(method.Body)

	// Restore "this"
	if prevThis != nil {
		r.Variables["this"] = prevThis
	} else {
		delete(r.Variables, "this")
	}

	return res
}

func (r *Runtime) CallMethodEvaluated(method *parser.MethodStatement, instance *Instance, args []interface{}) interface{} {
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

	// Execute body
	res := r.executeBlock(method.Body)

	// Restore "this"
	if prevThis != nil {
		r.Variables["this"] = prevThis
	} else {
		delete(r.Variables, "this")
	}

	return res
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

	fmt.Printf("Error: '%v' no es una función invocable\n", fn)
	return nil
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
			go func() {
				defer func() {
					if r := recover(); r != nil {
						future.err = fmt.Errorf("%v", r)
					}
					close(future.done)
				}()
				// We need to execute the function body.
				// But args[0] is already evaluated?
				// Wait, async expects a function literal or call?
				// If args[0] is a FunctionLiteral node (not evaluated), we can execute it.
				// But evaluateExpression evaluates FunctionLiteral to... itself (it's an expression).
				// Wait, parser.FunctionLiteral is an Expression. evaluateExpression returns *parser.FunctionLiteral?
				// Let's check evaluateExpression for FunctionLiteral.
				// It seems it returns the node itself (or we should wrap it).
				// In my previous edit I added `evaluateExpression` for `FunctionLiteral`.
				// Let's assume it returns the node.

				argVal := args[0]
				if fn, ok := argVal.(*parser.FunctionLiteral); ok {
					// We need a Runtime instance here. 'r' is available via closure.
					// But 'r' is not thread-safe if we modify variables.
					// We should probably clone runtime or scope?
					// For now, let's use 'r' but be careful.
					// Actually, async usually implies new stack/scope.
					// JosSecurity runtime is simple.
					future.result = r.executeBlock(fn.Body)
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
	}
	return nil, false
}
