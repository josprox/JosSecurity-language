package core

import (
	"fmt"
	"sort"

	semanticanalyzer "github.com/jossecurity/joss/pkg/analyzer"
	"github.com/jossecurity/joss/pkg/diagnostics"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

// AnalysisReport is the compatibility facade exposed by core. Diagnostics is
// canonical; Errors and Warnings remain populated for existing integrations.
type AnalysisReport struct {
	Diagnostics []diagnostics.Diagnostic
	Errors      []string
	Warnings    []string
}

func (ar *AnalysisReport) HasIssues() bool { return len(ar.Diagnostics) > 0 }
func (ar *AnalysisReport) HasErrors() bool { return len(ar.Errors) > 0 }

func (ar *AnalysisReport) PrintReport() {
	if !ar.HasIssues() {
		fmt.Println("Static analysis completed: no problems found.")
		return
	}
	fmt.Println("\nJoss static analysis")
	fmt.Println("------------------------------------------------------------")
	for _, item := range ar.Diagnostics {
		fmt.Println(item.String())
		if item.Suggestion != "" {
			fmt.Printf("  suggestion: %s\n", item.Suggestion)
		}
	}
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("%d error(s), %d warning(s)\n", len(ar.Errors), len(ar.Warnings))
}

// AnalyzeProgram analyzes an in-memory program. Project tooling should prefer
// AnalyzeSourceUnits so diagnostics retain their source file.
func AnalyzeProgram(program *parser.Program) *AnalysisReport {
	return AnalyzeSourceUnits([]semanticanalyzer.SourceUnit{{Path: "<memory>", Program: program}})
}

// AnalyzeSourceUnits performs project-aware, cross-file semantic analysis.
func AnalyzeSourceUnits(units []semanticanalyzer.SourceUnit) *AnalysisReport {
	environment := buildAnalysisEnvironment()
	items := semanticanalyzer.Analyze(units, environment)
	return AnalysisReportFromDiagnostics(items)
}

func AnalysisReportFromDiagnostics(items []diagnostics.Diagnostic) *AnalysisReport {
	report := &AnalysisReport{Diagnostics: items, Errors: []string{}, Warnings: []string{}}
	for _, item := range items {
		switch item.Severity {
		case diagnostics.SeverityError:
			report.Errors = append(report.Errors, item.String())
		case diagnostics.SeverityWarning:
			report.Warnings = append(report.Warnings, item.String())
		}
	}
	return report
}

// buildAnalysisEnvironment adapts the runtime's actual registries. It avoids a
// second hand-maintained class/function table: core natives and loaded plugin
// symbol indexes are the source of truth for the analyzer as well as execution.
func buildAnalysisEnvironment() semanticanalyzer.Environment {
	environment := semanticanalyzer.NewEnvironment()
	for _, name := range GetBuiltinFunctionNames() {
		environment.Builtins[name] = semanticanalyzer.Callable{
			Name: name, ReturnType: typesystem.Type{Kind: typesystem.Unknown}, Variadic: true,
		}
	}

	runtime := NewRuntime()
	defer runtime.Free()
	classNames := make([]string, 0, len(runtime.Classes))
	for name := range runtime.Classes {
		classNames = append(classNames, name)
	}
	sort.Strings(classNames)
	for _, name := range classNames {
		classNode := runtime.Classes[name]
		class := semanticanalyzer.Class{Name: name, Methods: make(map[string]semanticanalyzer.Callable)}
		if classNode != nil && classNode.SuperClass != nil {
			class.SuperClass = classNode.SuperClass.Value
		}
		if classNode != nil && classNode.Body != nil {
			for _, statement := range classNode.Body.Statements {
				switch method := statement.(type) {
				case *parser.MethodStatement:
					if method.Name != nil {
						class.Methods[method.Name.Value] = analysisCallable(method.Name.Value, method.Parameters, method.Body == nil)
					}
				case *parser.InitStatement:
					if method.Name != nil {
						class.Methods[method.Name.Value] = analysisCallable(method.Name.Value, method.Parameters, method.Body == nil)
					}
				}
			}
		}
		environment.Classes[name] = class
	}

	// Plugin symbol indexes carry signatures even for non-AST plugin formats.
	if runtime.PluginRegistry != nil {
		for _, plugin := range runtime.PluginRegistry.List() {
			for _, symbolClass := range plugin.Symbols.Classes {
				class := environment.Classes[symbolClass.Name]
				if class.Name == "" {
					class = semanticanalyzer.Class{Name: symbolClass.Name, Methods: make(map[string]semanticanalyzer.Callable)}
				}
				if class.Methods == nil {
					class.Methods = make(map[string]semanticanalyzer.Callable)
				}
				class.SuperClass = symbolClass.SuperClass
				for _, method := range symbolClass.Methods {
					parameters := make([]semanticanalyzer.Parameter, 0, len(method.Parameters))
					for _, parameter := range method.Parameters {
						parameterType := typesystem.Type{Kind: typesystem.Mixed}
						if parameter.Type != "" {
							parameterType = typesystem.Parse(parameter.Type)
						}
						parameters = append(parameters, semanticanalyzer.Parameter{Name: parameter.Name, Type: parameterType})
					}
					class.Methods[method.Name] = semanticanalyzer.Callable{Name: method.Name, Parameters: parameters, ReturnType: typesystem.Type{Kind: typesystem.Unknown}}
				}
				environment.Classes[class.Name] = class
			}
			for _, function := range plugin.Symbols.Functions {
				environment.Builtins[function.Name] = semanticanalyzer.Callable{Name: function.Name, ReturnType: typesystem.Type{Kind: typesystem.Unknown}, Variadic: true}
			}
		}
	}
	for name := range runtime.Variables {
		environment.Globals[name] = typesystem.Type{Kind: typesystem.Unknown}
	}
	return environment
}

func analysisCallable(name string, parameters []*parser.Parameter, signatureUnknown bool) semanticanalyzer.Callable {
	result := semanticanalyzer.Callable{Name: name, ReturnType: typesystem.Type{Kind: typesystem.Unknown}, Variadic: signatureUnknown}
	for _, parameter := range parameters {
		if parameter == nil || parameter.Name == nil {
			continue
		}
		parameterType := typesystem.Type{Kind: typesystem.Mixed}
		if parameter.Type.Literal != "" && parameter.Type.Type != parser.VAR {
			parameterType = typesystem.Parse(parameter.Type.Literal)
		}
		result.Parameters = append(result.Parameters, semanticanalyzer.Parameter{Name: parameter.Name.Value, Type: parameterType, HasDefault: parameter.DefaultValue != nil})
	}
	return result
}
