package core

import "fmt"

// VariableReference aliases one mutable variable binding for the duration of a
// Joss call. It never exposes a host address and is not a first-class Joss
// value.
type VariableReference struct {
	variables map[string]interface{}
	varTypes  map[string]string
	constants map[string]bool
	name      string
}

func (r *Runtime) referenceTo(name string) *VariableReference {
	if existing, ok := r.Variables[name].(*VariableReference); ok {
		return existing
	}
	return &VariableReference{variables: r.Variables, varTypes: r.VarTypes, constants: r.Constants, name: name}
}

func (reference *VariableReference) Get() interface{} {
	if reference == nil {
		return nil
	}
	value := reference.variables[reference.name]
	if nested, ok := value.(*VariableReference); ok {
		return nested.Get()
	}
	return value
}

func (reference *VariableReference) Type() string {
	if reference == nil {
		return ""
	}
	if nested, ok := reference.variables[reference.name].(*VariableReference); ok {
		return nested.Type()
	}
	return reference.varTypes[reference.name]
}

func (reference *VariableReference) Set(runtime *Runtime, value interface{}) interface{} {
	if reference == nil {
		panic("invalid nil Joss reference")
	}
	if nested, ok := reference.variables[reference.name].(*VariableReference); ok {
		return nested.Set(runtime, value)
	}
	if reference.constants[reference.name] {
		panic(&JossError{Type: "ConstantAssignment", Message: fmt.Sprintf("La constante '%s' no puede modificarse mediante ref", reference.name), File: runtime.CurrentFile})
	}
	expectedType := reference.varTypes[reference.name]
	if expectedType != "" && expectedType != "mixed" {
		value = runtime.coerceToTypedValue(value, expectedType)
		if !runtime.checkType(value, expectedType) {
			panic(&JossError{Type: "ReferenceTypeError", Message: fmt.Sprintf("La referencia '%s' requiere %s", reference.name, expectedType), File: runtime.CurrentFile})
		}
	}
	reference.variables[reference.name] = value
	return value
}
