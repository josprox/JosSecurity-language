package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) evaluateTernary(te *parser.TernaryExpression) interface{} {
	cond := r.evaluateExpression(te.Condition)
	isTrue := isTruthy(cond)

	var result interface{}

	if te.True == nil {
		// Elvis Operator
		if isTrue {
			result = cond
		} else {
			result = r.evaluateExpression(te.False)
		}
	} else {
		// Standard Ternary
		if isTrue {
			result = r.evaluateExpression(te.True)
		} else {
			result = r.evaluateExpression(te.False)
		}
	}

	if blk, ok := result.(*parser.BlockStatement); ok {
		return r.executeBlock(blk)
	}

	return result
}

func (r *Runtime) evaluateInfix(ie *parser.InfixExpression) interface{} {
	if ie.Operator == "??" {
		var left interface{}
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					left = nil
				}
			}()
			left = r.evaluateExpression(ie.Left)
		}()
		if left != nil {
			return left
		}
		return r.evaluateExpression(ie.Right)
	}

	left := r.evaluateExpression(ie.Left)

	// Short-Circuit Logic for && and ||
	if ie.Operator == "&&" {
		if !isTruthy(left) {
			return false
		}
		right := r.evaluateExpression(ie.Right)
		return isTruthy(right)
	}

	if ie.Operator == "||" {
		if isTruthy(left) {
			return true
		}
		right := r.evaluateExpression(ie.Right)
		return isTruthy(right)
	}

	// Handle cin >> $var
	if ie.Operator == ">>" {
		if _, ok := left.(*Cin); ok {
			if noInteract, ok := r.Env["NON_INTERACTIVE"]; ok && (noInteract == "true" || noInteract == "1") {
				fmt.Println("[Cin] Input skipped (NON_INTERACTIVE mode)")
				return nil
			}

			if ident, ok := ie.Right.(*parser.Identifier); ok {
				var input string
				fmt.Scanln(&input)

				var val interface{} = input
				if expectedType, exists := r.VarTypes[ident.Value]; exists {
					val = r.coerceToTypedValue(val, expectedType)
					if !r.checkType(val, expectedType) {
						fmt.Printf("Error de Tipado: No se puede asignar valor a '%s' (se espera %s)\n", ident.Value, expectedType)
						return nil
					}
				}

				r.Variables[ident.Value] = val
				return left
			}
			fmt.Println("Error: cin >> requiere una variable")
			return nil
		}
	}

	right := r.evaluateExpression(ie.Right)

	if ie.Operator == "===" {
		return strictCompare(left, right)
	}
	if ie.Operator == "!==" {
		return !strictCompare(left, right)
	}
	if ie.Operator == "<=>" {
		return spaceshipCompare(left, right)
	}

	// Handle cout << val or channel << val
	if ie.Operator == "<<" {
		if _, ok := left.(*Cout); ok {
			fmt.Print(right)
			return left
		}
		if ch, ok := left.(*Channel); ok {
			ch.Ch <- right
			return ch
		}
	}

	// Handle Pipe Operator |>
	if ie.Operator == "|>" {
		switch rightNode := ie.Right.(type) {
		case *parser.Identifier:
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
			var fn interface{}
			if ident, ok := rightNode.Function.(*parser.Identifier); ok {
				if f, ok := r.Functions[ident.Value]; ok {
					fn = f
				} else {
					args := []interface{}{left}
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

			args := []interface{}{left}
			for _, argExp := range rightNode.Arguments {
				args = append(args, r.evaluateExpression(argExp))
			}

			return r.applyFunction(fn, args)

		case *parser.FunctionLiteral:
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
		if i, ok := val.(int); ok {
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
		if ie.Operator == "/" {
			return lFloat / rFloat
		}
		if ie.Operator == "%" {
			return int64(lFloat) % int64(rFloat)
		}

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
			case "&&":
				return (lFloat != 0) && (rFloat != 0)
			case "||":
				return (lFloat != 0) || (rFloat != 0)
			}
		} else {
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
			case "%":
				return lInt % rInt
			case "&&":
				return (lInt != 0) && (rInt != 0)
			case "||":
				return (lInt != 0) || (rInt != 0)
			}
		}
	}

	lStr := ""
	rStr := ""
	if left != nil {
		lStr = fmt.Sprintf("%v", left)
	}
	if right != nil {
		rStr = fmt.Sprintf("%v", right)
	}

	if ie.Operator == "." {
		return lStr + rStr
	}
	if ie.Operator == "+" {
		fmt.Println("Error: El operador '+' es solo para números. Use '.' para concatenar cadenas.")
		return nil
	}
	if ie.Operator == "==" {
		return lStr == rStr
	}
	if ie.Operator == "!=" {
		return lStr != rStr
	}

	if bLeft, ok := left.(bool); ok {
		if bRight, ok := right.(bool); ok {
			if ie.Operator == "&&" {
				return bLeft && bRight
			}
			if ie.Operator == "||" {
				return bLeft || bRight
			}
		}
	}

	if ie.Operator == "??" {
		if left != nil {
			return left
		}
		return right
	}

	return nil
}

func (r *Runtime) evaluatePrefix(pe *parser.PrefixExpression) interface{} {
	right := r.evaluateExpression(pe.Right)

	if pe.Operator == "!" {
		return !isTruthy(right)
	}

	if pe.Operator == "-" {
		if i, ok := right.(int64); ok {
			return -i
		}
		if f, ok := right.(float64); ok {
			return -f
		}
	}

	return nil
}

func (r *Runtime) evaluatePostfix(pe *parser.PostfixExpression) interface{} {
	if pe.Operator == "++" {
		val := r.evaluateExpression(pe.Left)

		var newVal interface{}
		if i, ok := val.(int64); ok {
			newVal = i + 1
		} else if f, ok := val.(float64); ok {
			newVal = f + 1.0
		} else {
			fmt.Println("Error: Operador ++ solo aplicable a números")
			return nil
		}

		r.updateVariable(pe.Left, newVal)
		return val
	}
	return nil
}

func (r *Runtime) evaluateMatch(me *parser.MatchExpression) interface{} {
	subject := r.evaluateExpression(me.Subject)

	var defaultArm *parser.MatchArm
	for _, arm := range me.Arms {
		if arm.IsDefault {
			defaultArm = &arm
			continue
		}

		for _, keyExpr := range arm.Keys {
			keyVal := r.evaluateExpression(keyExpr)
			if strictCompare(subject, keyVal) {
				return r.evaluateExpression(arm.Value)
			}
		}
	}

	if defaultArm != nil {
		return r.evaluateExpression(defaultArm.Value)
	}

	return nil
}
