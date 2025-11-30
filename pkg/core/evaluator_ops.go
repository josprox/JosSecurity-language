package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

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

	// Handle Pipe Operator |>
	if ie.Operator == "|>" {
		// Right side can be:
		// 1. Identifier (function name) -> call(left)
		// 2. CallExpression (function call) -> call(left, args...)
		// 3. FunctionLiteral (anonymous function) -> call(left)

		switch rightNode := ie.Right.(type) {
		case *parser.Identifier:
			// Case 1: "hello" |> strtoupper
			fnName := rightNode.Value
			if fn, ok := r.Functions[fnName]; ok {
				return r.applyFunction(fn, []interface{}{left})
			}
			if res, ok := r.callBuiltin(fnName, []interface{}{left}); ok {
				return res
			}
			fmt.Printf("Error: Función '%s' no encontrada para pipe\n", fnName)
			return nil

		case *parser.CallExpression:
			// Case 2: "hello" |> foo(1) -> foo("hello", 1)

			// Evaluate function
			var fn interface{}
			if ident, ok := rightNode.Function.(*parser.Identifier); ok {
				if f, ok := r.Functions[ident.Value]; ok {
					fn = f
				} else {
					// Check builtin
					// But we need to evaluate args first to call builtin
					// Evaluate existing arguments
					args := []interface{}{left} // Prepend left
					for _, argExp := range rightNode.Arguments {
						args = append(args, r.evaluateExpression(argExp))
					}

					if res, ok := r.callBuiltin(ident.Value, args); ok {
						return res
					}
					fmt.Printf("Error: Función '%s' no encontrada en pipe call\n", ident.Value)
					return nil
				}
			} else {
				fn = r.evaluateExpression(rightNode.Function)
			}

			// Evaluate existing arguments
			args := []interface{}{left} // Prepend left
			for _, argExp := range rightNode.Arguments {
				args = append(args, r.evaluateExpression(argExp))
			}

			return r.applyFunction(fn, args)

		case *parser.FunctionLiteral:
			// Case 3: "hello" |> func($x) { return $x; }
			return r.applyFunction(rightNode, []interface{}{left})

		default:
			fmt.Printf("Error: El lado derecho del pipe debe ser una función o llamada, se obtuvo %T\n", ie.Right)
			return nil
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

	// Support Map access via dot notation (e.g. $item.id where $item is a map)
	if m, ok := left.(map[string]interface{}); ok {
		if val, exists := m[me.Property.Value]; exists {
			return val
		}
		return nil
	}

	instance, ok := left.(*Instance)
	if !ok {
		fmt.Printf("Error: %v (tipo %T) no es una instancia. Intentando acceder a: '%s'\n", left, left, me.Property.Value)
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
		if className == "Stack" || className == "Queue" || className == "GranMySQL" || className == "GranDB" || className == "Auth" ||
			className == "System" || className == "SmtpClient" || className == "Cron" || className == "Task" || className == "View" || className == "Router" ||
			className == "Request" || className == "Response" || className == "RedirectResponse" || className == "Session" || className == "Redirect" || className == "Security" || className == "Server" || className == "Log" ||
			className == "Schema" || className == "Blueprint" || className == "WebSocket" || className == "Redis" {
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

func (r *Runtime) evaluateEmpty(ee *parser.EmptyExpression) bool {
	// Special case: if argument is variable/index/member that doesn't exist, return true
	if !r.checkExistence(ee.Argument) {
		return true
	}

	val := r.evaluateExpression(ee.Argument)
	return isFalsy(val)
}
