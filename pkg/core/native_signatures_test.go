package core

import (
	"testing"

	"github.com/jossecurity/joss/pkg/typesystem"
)

func TestCoreCallableReturnSignaturesAreExplicit(t *testing.T) {
	environment := buildAnalysisEnvironment()
	for name := range builtinNamesMap {
		if environment.Builtins[name].ReturnType.Kind == typesystem.Unknown {
			t.Fatalf("builtin %s has unknown return type", name)
		}
	}
	runtime := NewRuntime()
	defer runtime.Free()
	for className, classNode := range runtime.Classes {
		if !IsNativeClass(className) || classNode == nil || classNode.Body == nil {
			continue
		}
		for methodName, callable := range environment.Classes[className].Methods {
			if callable.ReturnType.Kind == typesystem.Unknown {
				t.Fatalf("native %s::%s has unknown return type", className, methodName)
			}
		}
	}
}
