package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) evaluateExpression(exp parser.Expression) interface{} {
	switch e := exp.(type) {
	case *parser.StringLiteral:
		return e.Value
	case *parser.IntegerLiteral:
		return e.Value
	case *parser.FloatLiteral:
		return e.Value
	case *parser.Boolean:
		return e.Value
	case *parser.CallExpression:
		return r.executeCall(e)
	case *parser.Identifier:
		if val, ok := r.Variables[e.Value]; ok {
			return val
		}
		return nil
	case *parser.TernaryExpression:
		return r.evaluateTernary(e)
	case *parser.InfixExpression:
		return r.evaluateInfix(e)
	case *parser.ArrayLiteral:
		return r.evaluateArray(e)
	case *parser.MapLiteral:
		return r.evaluateMap(e)
	case *parser.IndexExpression:
		return r.evaluateIndex(e)
	case *parser.NewExpression:
		return r.evaluateNew(e)
	case *parser.MemberExpression:
		return r.evaluateMember(e)
	case *parser.AssignExpression:
		return r.evaluateAssign(e)
	case *parser.IssetExpression:
		return r.evaluateIsset(e)
	case *parser.EmptyExpression:
		return r.evaluateEmpty(e)
	case *parser.BlockExpression:
		// Return the block itself (or a closure wrapper if we had one)
		// For now, just return the BlockStatement so Task can execute it.
		return e.Block
	case *parser.FunctionLiteral:
		return e
	}
	return nil
}

func (r *Runtime) evaluateAssign(ae *parser.AssignExpression) interface{} {
	val := r.evaluateExpression(ae.Value)

	if ident, ok := ae.Left.(*parser.Identifier); ok {
		// Strict Typing Check
		if expectedType, exists := r.VarTypes[ident.Value]; exists {
			if !r.checkType(val, expectedType) {
				panic(fmt.Sprintf("Error de Tipado: No se puede asignar valor a '%s' (se espera %s)", ident.Value, expectedType))
			}
		}
		r.Variables[ident.Value] = val
		return val
	}

	if member, ok := ae.Left.(*parser.MemberExpression); ok {
		left := r.evaluateExpression(member.Left)
		if instance, ok := left.(*Instance); ok {
			instance.Fields[member.Property.Value] = val
			return val
		}
		fmt.Printf("Error: Asignación a miembro de no-instancia: %v\n", left)
		return nil
	}

	fmt.Printf("Error: Asignación inválida a %T\n", ae.Left)
	return nil
}

func (r *Runtime) evaluateArray(al *parser.ArrayLiteral) []interface{} {
	elements := []interface{}{}
	for _, el := range al.Elements {
		elements = append(elements, r.evaluateExpression(el))
	}
	return elements
}

func (r *Runtime) evaluateMap(ml *parser.MapLiteral) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range ml.Pairs {
		key := r.evaluateExpression(k)
		val := r.evaluateExpression(v)
		if keyStr, ok := key.(string); ok {
			m[keyStr] = val
		} else {
			fmt.Printf("Error: Clave de mapa inválida: %v (se espera string)\n", key)
		}
	}
	return m
}

func (r *Runtime) evaluateIndex(ie *parser.IndexExpression) interface{} {
	left := r.evaluateExpression(ie.Left)
	index := r.evaluateExpression(ie.Index)

	if list, ok := left.([]interface{}); ok {
		if idx, ok := index.(int64); ok {
			if idx >= 0 && idx < int64(len(list)) {
				return list[idx]
			}
			fmt.Println("Error: Índice fuera de rango")
		} else {
			fmt.Println("Error: El índice debe ser un entero")
		}
		return nil
	}

	if m, ok := left.(map[string]interface{}); ok {
		if key, ok := index.(string); ok {
			if val, exists := m[key]; exists {
				return val
			}
			return nil
		}
		fmt.Println("Error: El índice de un mapa debe ser string")
		return nil
	}

	fmt.Println("Error: No se puede indexar algo que no es un array o mapa")
	return nil
}

func (r *Runtime) evaluateTernary(te *parser.TernaryExpression) interface{} {
	cond := r.evaluateExpression(te.Condition)
	isTrue := false

	if b, ok := cond.(bool); ok {
		isTrue = b
	} else if s, ok := cond.(string); ok {
		isTrue = s == "true" || s == "TRUE"
	}

	if isTrue {
		if te.True != nil {
			return r.executeBlock(te.True)
		}
	} else {
		if te.False != nil {
			return r.executeBlock(te.False)
		}
	}
	return nil
}

func (r *Runtime) evaluateInfix(ie *parser.InfixExpression) interface{} {
	left := r.evaluateExpression(ie.Left)

	// Handle cin >> $var (Special case: Right is not evaluated as expression, but as l-value)
	if ie.Operator == ">>" {
		if _, ok := left.(*Cin); ok {
			if ident, ok := ie.Right.(*parser.Identifier); ok {
				var input string
				fmt.Scanln(&input)
				r.Variables[ident.Value] = input
				return left // Return cin for chaining?
			}
			fmt.Println("Error: cin >> requiere una variable")
			return nil
		}
	}

	right := r.evaluateExpression(ie.Right)

	// Handle cout << val or channel << val
	if ie.Operator == "<<" {
		if _, ok := left.(*Cout); ok {
			fmt.Print(right)
			return left // Return cout for chaining
		}
		if ch, ok := left.(*Channel); ok {
			ch.Ch <- right
			return ch // Return channel for chaining?
		}
	}

	// Smart Numerics: Auto-promote to float if needed
	toFloat := func(val interface{}) (float64, bool) {
		if i, ok := val.(int64); ok {
			return float64(i), true
		}
		if f, ok := val.(float64); ok {
			return f, true
		}
		return 0, false
	}

	lFloat, lIsNum := toFloat(left)
	rFloat, rIsNum := toFloat(right)

	if lIsNum && rIsNum {
		// If division, always float
		if ie.Operator == "/" {
			return lFloat / rFloat
		}

		// If any operand is float, result is float
		isFloatOp := false
		if _, ok := left.(float64); ok {
			isFloatOp = true
		}
		if _, ok := right.(float64); ok {
			isFloatOp = true
		}

		if isFloatOp {
			switch ie.Operator {
			case "+":
				return lFloat + rFloat
			case "-":
				return lFloat - rFloat
			case "*":
				return lFloat * rFloat
			case "<":
				return lFloat < rFloat
			case ">":
				return lFloat > rFloat
			case ">=":
				return lFloat >= rFloat
			case "<=":
				return lFloat <= rFloat
			case "==":
				return lFloat == rFloat
			case "!=":
				return lFloat != rFloat
			}
		} else {
			// Integer operations
			lInt := int64(lFloat)
			rInt := int64(rFloat)
			switch ie.Operator {
			case "+":
				return lInt + rInt
			case "-":
				return lInt - rInt
			case "*":
				return lInt * rInt
			case "<":
				return lInt < rInt
			case ">":
				return lInt > rInt
			case ">=":
				return lInt >= rInt
			case "<=":
				return lInt <= rInt
			case "==":
				return lInt == rInt
			case "!=":
				return lInt != rInt
			}
		}
	}

	lStr := fmt.Sprintf("%v", left)
	rStr := fmt.Sprintf("%v", right)
	if ie.Operator == "+" {
		return lStr + rStr
	}
	if ie.Operator == "==" {
		return lStr == rStr
	}
	if ie.Operator == "!=" {
		return lStr != rStr
	}

	return nil
}

func (r *Runtime) evaluateNew(ne *parser.NewExpression) interface{} {
	className := ne.Class.Value
	classStmt, ok := r.Classes[className]
	if !ok {
		fmt.Printf("Error: Clase '%s' no encontrada\n", className)
		return nil
	}

	instance := &Instance{
		Class:  classStmt,
		Fields: make(map[string]interface{}),
	}

	// Collect inheritance chain
	chain := []*parser.ClassStatement{classStmt}
	curr := classStmt
	for curr.SuperClass != nil {
		parentName := curr.SuperClass.Value
		if parent, ok := r.Classes[parentName]; ok {
			chain = append(chain, parent)
			curr = parent
		} else {
			break
		}
	}

	// Initialize properties (Parent -> Child)
	for i := len(chain) - 1; i >= 0; i-- {
		cls := chain[i]
		for _, stmt := range cls.Body.Statements {
			if let, ok := stmt.(*parser.LetStatement); ok {
				instance.Fields[let.Name.Value] = r.evaluateExpression(let.Value)
			}
		}
	}

	// Call constructor if exists
	for _, stmt := range classStmt.Body.Statements {
		if method, ok := stmt.(*parser.MethodStatement); ok {
			if method.Name.Value == "constructor" {
				r.CallMethod(method, instance, ne.Arguments)
				break
			}
		}
		if initStmt, ok := stmt.(*parser.InitStatement); ok {
			if initStmt.Name.Value == "constructor" {
				// Convert to MethodStatement
				method := &parser.MethodStatement{
					Token:      initStmt.Token,
					Name:       initStmt.Name,
					Parameters: initStmt.Parameters,
					Body:       initStmt.Body,
				}
				r.CallMethod(method, instance, ne.Arguments)
				break
			}
		}
	}

	return instance
}

func (r *Runtime) evaluateMember(me *parser.MemberExpression) interface{} {
	left := r.evaluateExpression(me.Left)
	instance, ok := left.(*Instance)
	if !ok {
		fmt.Printf("Error: %v no es una instancia\n", left)
		return nil
	}

	propName := me.Property.Value

	// Check fields
	if val, ok := instance.Fields[propName]; ok {
		return val
	}

	// Check methods (Function and Init)
	currentClass := instance.Class
	for currentClass != nil {
		for _, stmt := range currentClass.Body.Statements {
			if method, ok := stmt.(*parser.MethodStatement); ok {
				if method.Name.Value == propName {
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
			if initStmt, ok := stmt.(*parser.InitStatement); ok {
				if initStmt.Name.Value == propName {
					// Convert InitStatement to MethodStatement for compatibility
					method := &parser.MethodStatement{
						Token:      initStmt.Token,
						Name:       initStmt.Name,
						Parameters: initStmt.Parameters,
						Body:       initStmt.Body,
					}
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
		}

		// Move to parent
		if currentClass.SuperClass != nil {
			parentName := currentClass.SuperClass.Value
			if parent, ok := r.Classes[parentName]; ok {
				currentClass = parent
			} else {
				// fmt.Printf("Error: Clase padre '%s' no encontrada\n", parentName)
				currentClass = nil
			}
		} else {
			currentClass = nil
		}
	}

	// Check for Native Class methods
	checkClass := instance.Class
	isNative := false
	for checkClass != nil {
		className := checkClass.Name.Value
		if className == "Stack" || className == "Queue" || className == "GranMySQL" || className == "Auth" ||
			className == "System" || className == "SmtpClient" || className == "Cron" || className == "Task" || className == "View" || className == "Router" ||
			className == "Request" || className == "Response" || className == "RedirectResponse" || className == "Session" || className == "Redirect" || className == "Security" || className == "Server" || className == "Log" {
			isNative = true
			break
		}
		if checkClass.SuperClass != nil {
			if parent, ok := r.Classes[checkClass.SuperClass.Value]; ok {
				checkClass = parent
			} else {
				break
			}
		} else {
			break
		}
	}

	if isNative {
		// Return a synthetic method with nil body to trigger native execution
		return &BoundMethod{
			Method: &parser.MethodStatement{
				Name: &parser.Identifier{Value: propName},
				Body: nil,
			},
			Instance: instance,
		}
	}

	fmt.Printf("Error: Propiedad o método '%s' no encontrado\n", propName)
	return nil
}

func (r *Runtime) evaluateIsset(ie *parser.IssetExpression) bool {
	for _, arg := range ie.Arguments {
		if !r.checkExistence(arg) {
			return false
		}
	}
	return true
}

func (r *Runtime) checkExistence(exp parser.Expression) bool {
	switch e := exp.(type) {
	case *parser.Identifier:
		_, ok := r.Variables[e.Value]
		return ok
	case *parser.IndexExpression:
		left := r.evaluateExpression(e.Left)
		if list, ok := left.([]interface{}); ok {
			index := r.evaluateExpression(e.Index)
			if idx, ok := index.(int64); ok {
				return idx >= 0 && idx < int64(len(list))
			}
		}
		return false
	case *parser.MemberExpression:
		left := r.evaluateExpression(e.Left)
		if instance, ok := left.(*Instance); ok {
			_, ok := instance.Fields[e.Property.Value]
			return ok
		}
		return false
	}
	return false
}

func (r *Runtime) evaluateEmpty(ee *parser.EmptyExpression) bool {
	// Special case: if argument is variable/index/member that doesn't exist, return true
	if !r.checkExistence(ee.Argument) {
		return true
	}

	val := r.evaluateExpression(ee.Argument)
	return isFalsy(val)
}

func isFalsy(val interface{}) bool {
	if val == nil {
		return true
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

func (r *Runtime) checkType(val interface{}, typeName string) bool {
	if val == nil {
		return true
	} // Allow nil?
	switch typeName {
	case "int":
		_, ok := val.(int64)
		return ok
	case "float":
		_, ok := val.(float64)
		return ok
	case "string":
		_, ok := val.(string)
		return ok
	case "bool":
		_, ok := val.(bool)
		return ok
	case "array":
		_, ok := val.([]interface{})
		return ok
	case "map":
		_, ok := val.(map[string]interface{})
		return ok
	case "channel":
		_, ok := val.(*Channel)
		return ok
	}
	return true // Unknown types (classes) not strictly checked yet
}
