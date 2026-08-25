package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

func (r *Runtime) evaluateAssign(ae *parser.AssignExpression) interface{} {
	val := r.evaluateExpression(ae.Value)

	if ident, ok := ae.Left.(*parser.Identifier); ok {
		if expectedType, exists := r.VarTypes[ident.Value]; exists {
			if expectedType != "mixed" {
				val = r.coerceToTypedValue(val, expectedType)
				if !r.checkType(val, expectedType) {
					panic(fmt.Sprintf("Error de Tipado: No se puede asignar valor a '%s' (se espera %s)", ident.Value, expectedType))
				}
			}
		} else if _, alreadyExists := r.Variables[ident.Value]; !alreadyExists || r.Variables[ident.Value] == nil {
			if inferredType := runtimeTypeName(val); inferredType != "" {
				r.VarTypes[ident.Value] = inferredType
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

	if indexExp, ok := ae.Left.(*parser.IndexExpression); ok {
		left := r.evaluateExpression(indexExp.Left)

		if indexExp.Index == nil {
			if list, ok := left.([]interface{}); ok {
				newList := append(list, val)
				return r.updateVariable(indexExp.Left, newList)
			}
			fmt.Println("Error: Append [] solo permitido en arrays")
			return nil
		}

		index := r.evaluateExpression(indexExp.Index)

		if m, ok := left.(map[string]interface{}); ok {
			if key, ok := index.(string); ok {
				m[key] = val
				return val
			}
		}

		if list, ok := left.([]interface{}); ok {
			if idx, ok := index.(int64); ok {
				if idx >= 0 && idx < int64(len(list)) {
					list[idx] = val
					return val
				}
			}
		}
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

	if str, ok := left.(string); ok {
		if idx, ok := index.(int64); ok {
			if idx >= 0 && idx < int64(len(str)) {
				return string(str[idx])
			}
			fmt.Println("Error: Índice de string fuera de rango")
			return nil
		}
	}

	fmt.Println("Error: No se puede indexar algo que no es un array o mapa")
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
	if !r.checkExistence(ee.Argument) {
		return true
	}

	val := r.evaluateExpression(ee.Argument)
	return isFalsy(val)
}

func (r *Runtime) updateVariable(exp parser.Expression, newVal interface{}) interface{} {
	if ident, ok := exp.(*parser.Identifier); ok {
		r.Variables[ident.Value] = newVal
		return newVal
	}
	if member, ok := exp.(*parser.MemberExpression); ok {
		left := r.evaluateExpression(member.Left)
		if instance, ok := left.(*Instance); ok {
			instance.Fields[member.Property.Value] = newVal
			return newVal
		}
	}
	fmt.Println("Error: No se puede actualizar la variable (expresión no soportada)")
	return nil
}
