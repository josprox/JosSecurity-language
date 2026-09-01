package core

import "github.com/jossecurity/joss/pkg/parser"

// captureFunction snapshots the current lexical environment for a callback
// that will run after the registering method has returned.
func (r *Runtime) captureFunction(fn *parser.FunctionLiteral) *CapturedFunction {
	variables := make(map[string]interface{}, len(r.Variables))
	for name, value := range r.Variables {
		if r.sourceMapVisible(name) {
			variables[name] = value
		}
	}

	varTypes := make(map[string]string, len(r.VarTypes))
	for name, valueType := range r.VarTypes {
		if r.sourceMapVisible(name) {
			varTypes[name] = valueType
		}
	}
	constants := make(map[string]bool, len(r.Constants))
	for name, constant := range r.Constants {
		if constant && r.sourceMapVisible(name) {
			constants[name] = true
		}
	}
	if r.currentFrame != nil {
		for index := range r.currentFrame.slots {
			slot := &r.currentFrame.slots[index]
			if !slot.Initialized {
				continue
			}
			val := slot.Value.Interface()
			variables[slot.Name] = val
			variables["$"+slot.Name] = val
			varTypes[slot.Name] = slot.TypeName
			varTypes["$"+slot.Name] = slot.TypeName
			if slot.Constant {
				constants[slot.Name] = true
				constants["$"+slot.Name] = true
			}
		}
	}

	captureEnv := &ClosureEnvironment{
		Variables: variables,
		VarTypes:  varTypes,
		Constants: constants,
	}

	return &CapturedFunction{
		Function:    fn,
		Environment: captureEnv,
	}
}

func (r *Runtime) callCapturedFunction(closure *CapturedFunction, args []interface{}) interface{} {
	environment := closure.Environment
	environment.mu.Lock()
	defer environment.mu.Unlock()

	previousVariables := r.Variables
	previousVarTypes := r.VarTypes
	previousConstants := r.Constants
	r.Variables = environment.Variables
	r.VarTypes = environment.VarTypes
	r.Constants = environment.Constants
	defer func() {
		r.Variables = previousVariables
		r.VarTypes = previousVarTypes
		r.Constants = previousConstants
	}()

	method := &parser.MethodStatement{
		Token:      closure.Function.Token,
		Name:       &parser.Identifier{Value: "anonymous"},
		Parameters: closure.Function.Parameters,
		ReturnType: closure.Function.ReturnType,
		Body:       closure.Function.Body,
	}
	return r.callMethodEvaluatedWithPlan(method, nil, args, environment, r.planForFunction(closure.Function))
}
