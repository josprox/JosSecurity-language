package core

import "github.com/jossecurity/joss/pkg/parser"

// captureFunction snapshots the current lexical environment for a callback
// that will run after the registering method has returned.
func (r *Runtime) captureFunction(fn *parser.FunctionLiteral) *CapturedFunction {
	if r.captureEnvironment == nil {
		variables := make(map[string]interface{}, len(r.Variables))
		for name, value := range r.Variables {
			variables[name] = value
		}

		varTypes := make(map[string]string, len(r.VarTypes))
		for name, valueType := range r.VarTypes {
			varTypes[name] = valueType
		}

		r.captureEnvironment = &ClosureEnvironment{
			Variables: variables,
			VarTypes:  varTypes,
		}
	}

	return &CapturedFunction{
		Function:    fn,
		Environment: r.captureEnvironment,
	}
}

func (r *Runtime) callCapturedFunction(closure *CapturedFunction, args []interface{}) interface{} {
	environment := closure.Environment
	environment.mu.Lock()
	defer environment.mu.Unlock()

	previousVariables := r.Variables
	previousVarTypes := r.VarTypes
	r.Variables = environment.Variables
	r.VarTypes = environment.VarTypes
	defer func() {
		r.Variables = previousVariables
		r.VarTypes = previousVarTypes
	}()

	method := &parser.MethodStatement{
		Token:      closure.Function.Token,
		Name:       &parser.Identifier{Value: "anonymous"},
		Parameters: closure.Function.Parameters,
		Body:       closure.Function.Body,
	}
	return r.CallMethodEvaluated(method, nil, args)
}
