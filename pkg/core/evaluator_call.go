package core

import (
	"fmt"
	"reflect"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) CallMethod(method *parser.MethodStatement, instance *Instance, args []parser.Expression) (res interface{}) {
	evaluated := make([]interface{}, 0, len(args))
	for _, argument := range args {
		evaluated = append(evaluated, r.evaluateExpression(argument))
	}
	return r.CallMethodEvaluated(method, instance, evaluated)
}

func (r *Runtime) CallMethodEvaluated(method *parser.MethodStatement, instance *Instance, args []interface{}) (res interface{}) {
	return r.callMethodEvaluated(method, instance, args, nil)
}

func (r *Runtime) callMethodEvaluated(method *parser.MethodStatement, instance *Instance, args []interface{}, writeBack *ClosureEnvironment) (res interface{}) {
	// Native Method Support
	if method.Body == nil {
		return r.executeNativeMethod(instance, method.Name.Value, args)
	}
	required := 0
	for _, parameter := range method.Parameters {
		if parameter.DefaultValue == nil {
			required++
		}
	}
	if len(args) < required || len(args) > len(method.Parameters) {
		panic(fmt.Sprintf("Arity Error: %s() espera entre %d y %d argumentos, se recibieron %d", method.Name.Value, required, len(method.Parameters), len(args)))
	}
	if r.MaxCallDepth <= 0 {
		r.MaxCallDepth = DefaultMaxCallDepth
	}
	if r.callDepth >= r.MaxCallDepth {
		panic(&JossError{
			Type:    "RecursionLimit",
			Message: fmt.Sprintf("La llamada a '%s' excedió el límite de recursión de %d frames", method.Name.Value, r.MaxCallDepth),
			File:    r.CurrentFile,
			Line:    method.Token.Line,
		})
	}
	r.callDepth++
	defer func() { r.callDepth-- }()

	// Named callables see only runtime/plugin globals. Function literals receive
	// their captured lexical environment through writeBack. This prevents the
	// accidental dynamic scoping that used to expose caller locals to recursion.
	parentVariables := r.Variables
	parentVarTypes := r.VarTypes
	parentConstants := r.Constants
	frameVariables := make(map[string]interface{}, len(r.HostGlobals)+len(method.Parameters)+1)
	frameVarTypes := make(map[string]string, len(parentVarTypes)+len(method.Parameters))
	frameConstants := make(map[string]bool)
	for name, value := range parentVariables {
		if writeBack != nil || r.HostGlobals[name] {
			frameVariables[name] = value
			if valueType, exists := parentVarTypes[name]; exists {
				frameVarTypes[name] = valueType
			}
			if parentConstants[name] {
				frameConstants[name] = true
			}
		}
	}
	r.Variables = frameVariables
	r.VarTypes = frameVarTypes
	r.Constants = frameConstants
	parameterNames := make(map[string]bool, len(method.Parameters))
	for _, parameter := range method.Parameters {
		parameterNames[parameter.Name.Value] = true
	}
	defer func() {
		if writeBack != nil {
			for name := range writeBack.Variables {
				if !parameterNames[name] {
					writeBack.Variables[name] = r.Variables[name]
				}
			}
			for name := range writeBack.VarTypes {
				if !parameterNames[name] {
					writeBack.VarTypes[name] = r.VarTypes[name]
				}
			}
			writeBack.Constants = copyBoolMap(r.Constants)
		}
		r.Variables = parentVariables
		r.VarTypes = parentVarTypes
		r.Constants = parentConstants
	}()

	if instance != nil {
		r.Variables["this"] = instance
	} else {
		delete(r.Variables, "this")
		delete(r.VarTypes, "this")
	}
	previousCaptureEnvironment := r.captureEnvironment
	r.captureEnvironment = nil
	defer func() { r.captureEnvironment = previousCaptureEnvironment }()

	// Bind arguments
	for i, param := range method.Parameters {
		declaredType := param.Type.Literal
		if declaredType == "" {
			declaredType = "mixed"
		}
		r.VarTypes[param.Name.Value] = declaredType
		var val interface{}
		if i < len(args) {
			val = args[i]
			if param.Type.Literal != "" {
				val = r.coerceToTypedValue(val, param.Type.Literal)
				if !r.checkType(val, param.Type.Literal) {
					panic(fmt.Sprintf("Type Error: El argumento %d ($%s) debe ser de tipo %s, se recibió %T", i+1, param.Name.Value, param.Type.Literal, val))
				}
			}
		} else if param.DefaultValue != nil {
			val = r.evaluateExpression(param.DefaultValue)
			if param.Type.Literal != "" {
				val = r.coerceToTypedValue(val, param.Type.Literal)
				if !r.checkType(val, param.Type.Literal) {
					panic(fmt.Sprintf("Type Error: El valor por defecto de $%s debe ser de tipo %s, se recibió %T", param.Name.Value, param.Type.Literal, val))
				}
			}
		} else {
			val = nil
		}
		r.Variables[param.Name.Value] = val
	}

	defer func() {
		if p := recover(); p != nil {
			if rp, ok := p.(*ReturnPanic); ok {
				res = rp.Value
			} else {
				panic(p)
			}
		}
		if method.ReturnType.Literal != "" {
			res = r.coerceToTypedValue(res, method.ReturnType.Literal)
			if !r.checkType(res, method.ReturnType.Literal) {
				panic(&JossError{
					Type:    "ReturnTypeError",
					Message: fmt.Sprintf("La función '%s' debe retornar %s, recibió %T", method.Name.Value, method.ReturnType.Literal, res),
					File:    r.CurrentFile,
					Line:    method.Token.Line,
				})
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

	// 3. For plain identifiers: check r.Functions first, then r.Variables.
	//    This avoids a false "UndefinedVariable" panic for user-defined functions.
	var fn interface{}
	if ident, ok := call.Function.(*parser.Identifier); ok {
		if f, ok := r.Functions[ident.Value]; ok {
			fn = f
		} else if v, ok := r.Variables[ident.Value]; ok {
			fn = v
		}
	} else {
		// Non-identifier (e.g., member expression, closure variable): evaluate normally.
		fn = r.evaluateExpression(call.Function)
	}

	if fn == nil {
		if ident, ok := call.Function.(*parser.Identifier); ok {
			panic(&JossError{
				Type:    "UndefinedFunction",
				Message: fmt.Sprintf("Función '%s' no encontrada", ident.Value),
				File:    r.CurrentFile,
				Line:    ident.Token.Line,
			})
		}
		panic(&JossError{
			Type:    "NotCallable",
			Message: "Intento de invocar un objeto nulo o no invocable",
			File:    r.CurrentFile,
		})
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
			evalArgs := []interface{}{}
			for _, arg := range args {
				evalArgs = append(evalArgs, arg)
			}
			classStmt := r.Classes[bound.StaticClass]
			if classStmt == nil {
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
		method := &parser.MethodStatement{
			Token:      lit.Token,
			Name:       &parser.Identifier{Value: "anonymous"},
			Parameters: lit.Parameters,
			ReturnType: lit.ReturnType,
			Body:       lit.Body,
		}
		return r.CallMethodEvaluated(method, nil, args)
	}

	if handler, ok := fn.(NativeHandler); ok {
		return handler(r, nil, "", args)
	}

	if goFn, ok := fn.(func([]interface{}) interface{}); ok {
		return goFn(args)
	}

	val := reflect.ValueOf(fn)
	if val.Kind() == reflect.Func {
		fnType := val.Type()
		inArgs := []reflect.Value{}
		for i := 0; i < fnType.NumIn() && i < len(args); i++ {
			expectedType := fnType.In(i)
			if args[i] == nil {
				inArgs = append(inArgs, reflect.Zero(expectedType))
			} else {
				argVal := reflect.ValueOf(args[i])
				if argVal.Type().AssignableTo(expectedType) {
					inArgs = append(inArgs, argVal)
				} else if argVal.Type().ConvertibleTo(expectedType) {
					inArgs = append(inArgs, argVal.Convert(expectedType))
				} else {
					inArgs = append(inArgs, reflect.Zero(expectedType))
				}
			}
		}
		results := val.Call(inArgs)
		if len(results) > 0 {
			return results[0].Interface()
		}
		return nil
	}

	if fn == nil {
		panic(&JossError{
			Type:    "NotCallable",
			Message: "Intento de invocar un valor nulo",
			File:    r.CurrentFile,
		})
	}

	panic(&JossError{
		Type:    "NotCallable",
		Message: fmt.Sprintf("'%v' (tipo %T) no es una función invocable", fn, fn),
		File:    r.CurrentFile,
	})
}

// CallFunction is the public API for executing functions
func (r *Runtime) CallFunction(fn interface{}, args []interface{}) interface{} {
	return r.applyFunction(fn, args)
}

// callBuiltin dispatches to domain-specific built-in handlers
func (r *Runtime) callBuiltin(name string, args []interface{}) (interface{}, bool) {
	if !IsBuiltin(name) {
		return nil, false
	}
	// 1. Date & Time Builtins
	if res, ok := r.callBuiltinDate(name, args); ok {
		return res, true
	}

	// 2. String & Formatting Builtins
	if res, ok := r.callBuiltinString(name, args); ok {
		return res, true
	}

	// 3. Array & Map Builtins
	if res, ok := r.callBuiltinArray(name, args); ok {
		return res, true
	}

	// 4. Async & Channels Builtins
	if res, ok := r.callBuiltinAsync(name, args); ok {
		return res, true
	}

	// 5. IO, Web & Formats Builtins
	if res, ok := r.callBuiltinIO(name, args); ok {
		return res, true
	}

	panic(fmt.Sprintf("internal error: builtin %q is catalogued without a runtime handler", name))
}
