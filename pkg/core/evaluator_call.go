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

func (r *Runtime) executeCall(call *parser.CallExpression) interface{} {
	fn := r.evaluateExpression(call.Function)

	if bound, ok := fn.(*BoundMethod); ok {
		return r.CallMethod(bound.Method, bound.Instance, call.Arguments)
	}

	if ident, ok := call.Function.(*parser.Identifier); ok {
		// Check for Global Function
		if method, ok := r.Functions[ident.Value]; ok {
			return r.CallMethod(method, nil, call.Arguments)
		}

		if ident.Value == "print" || ident.Value == "echo" {
			for _, arg := range call.Arguments {
				val := r.evaluateExpression(arg)
				fmt.Println(val)
			}
			return nil
		}

		if ident.Value == "printf" {
			if len(call.Arguments) > 0 {
				format := r.evaluateExpression(call.Arguments[0])
				args := []interface{}{}
				for _, arg := range call.Arguments[1:] {
					args = append(args, r.evaluateExpression(arg))
				}
				if fmtStr, ok := format.(string); ok {
					fmt.Printf(fmtStr, args...)
				}
			}
			return nil
		}
		if ident.Value == "env" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if key, ok := arg.(string); ok {
					if val, exists := r.Env[key]; exists {
						return val
					}
					return ""
				}
			}
			return nil
		}
		if ident.Value == "len" || ident.Value == "count" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if list, ok := arg.([]interface{}); ok {
					return int64(len(list))
				}
				if str, ok := arg.(string); ok {
					return int64(len(str))
				}
			}
			return int64(0)
		}
		if ident.Value == "toon_encode" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				return ToonEncode(arg)
			}
			return ""
		}
		if ident.Value == "toon_decode" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if str, ok := arg.(string); ok {
					return ToonDecode(str)
				}
			}
			return nil
		}

		if ident.Value == "toon_verify" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if str, ok := arg.(string); ok {
					return ToonVerify(str)
				}
			}
			return false
		}
		if ident.Value == "json_encode" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				return JsonEncode(arg)
			}
			return ""
		}
		if ident.Value == "json_decode" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if str, ok := arg.(string); ok {
					return JsonDecode(str)
				}
			}
			return nil
		}
		if ident.Value == "json_verify" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if str, ok := arg.(string); ok {
					return JsonVerify(str)
				}
			}
			return false
		}
		// async function - executes code in goroutine
		if ident.Value == "async" {
			if len(call.Arguments) == 1 {
				future := &Future{
					done: make(chan bool),
				}

				// Execute the argument in a goroutine
				go func() {
					defer func() {
						if r := recover(); r != nil {
							future.err = fmt.Errorf("%v", r)
						}
						close(future.done)
					}()

					// Evaluate the argument
					argVal := r.evaluateExpression(call.Arguments[0])
					if fn, ok := argVal.(*parser.FunctionLiteral); ok {
						// Execute function body
						future.result = r.executeBlock(fn.Body)
					} else {
						future.result = argVal
					}
				}()

				return future
			}
			return nil
		}
		// await function - waits for a Future
		if ident.Value == "await" {
			if len(call.Arguments) == 1 {
				futureVal := r.evaluateExpression(call.Arguments[0])
				if future, ok := futureVal.(*Future); ok {
					return future.Wait()
				}
				fmt.Println("Error: await expects a Future")
			}
			return nil
		}

		// Channel Functions
		if ident.Value == "make_chan" {
			size := 0
			if len(call.Arguments) > 0 {
				if s, ok := r.evaluateExpression(call.Arguments[0]).(int64); ok {
					size = int(s)
				}
			}
			return &Channel{Ch: make(chan interface{}, size)}
		}

		if ident.Value == "close" {
			if len(call.Arguments) == 1 {
				if ch, ok := r.evaluateExpression(call.Arguments[0]).(*Channel); ok {
					close(ch.Ch)
					return nil
				}
				fmt.Println("Error: close expects a channel")
			}
			return nil
		}

		if ident.Value == "send" {
			if len(call.Arguments) == 2 {
				chVal := r.evaluateExpression(call.Arguments[0])
				val := r.evaluateExpression(call.Arguments[1])
				if ch, ok := chVal.(*Channel); ok {
					ch.Ch <- val
					return nil
				}
				fmt.Println("Error: send expects (channel, value)")
			}
			return nil
		}

		if ident.Value == "recv" {
			if len(call.Arguments) == 1 {
				if ch, ok := r.evaluateExpression(call.Arguments[0]).(*Channel); ok {
					val, ok := <-ch.Ch
					if !ok {
						return nil // Channel closed
					}
					return val
				}
				fmt.Println("Error: recv expects a channel")
			}
			return nil
		}
	}
	return nil
}
