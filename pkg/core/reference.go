package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
	runtimeframe "github.com/jossecurity/joss/pkg/runtime/frame"
)

// VariableReference aliases one mutable variable binding for the duration of a
// Joss call. It never exposes a host address and is not a first-class Joss
// value.
type VariableReference struct {
	variables map[string]interface{}
	varTypes  map[string]string
	constants map[string]bool
	name      string
	slot      *runtimeframe.Slot
}

func (r *Runtime) referenceToIdentifier(identifier *parser.Identifier) *VariableReference {
	if slot, resolved := r.slotForIdentifier(identifier); resolved {
		if existing, ok := slot.Value.Interface().(*VariableReference); ok && slot.Initialized {
			return existing
		}
		return &VariableReference{name: identifier.Value, slot: slot}
	}
	return r.referenceTo(identifier.Value)
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
	if reference.slot != nil {
		value := reference.slot.Value.Interface()
		if nested, ok := value.(*VariableReference); ok {
			return nested.Get()
		}
		return value
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
	if reference.slot != nil {
		if nested, ok := reference.slot.Value.Interface().(*VariableReference); ok {
			return nested.Type()
		}
		return reference.slot.TypeName
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
	if reference.slot != nil {
		if nested, ok := reference.slot.Value.Interface().(*VariableReference); ok {
			return nested.Set(runtime, value)
		}
		if reference.slot.Constant {
			panic(&JossError{Type: "ConstantAssignment", Message: fmt.Sprintf("La constante '%s' no puede modificarse mediante ref", reference.name), File: runtime.CurrentFile})
		}
		return runtime.assignSlot(reference.slot, value, false)
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
