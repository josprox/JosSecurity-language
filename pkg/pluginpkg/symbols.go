package pluginpkg

import "github.com/jossecurity/joss/pkg/parser"

const SymbolSchemaVersion = 1

type SymbolIndex struct {
	Schema    int                 `json:"schema"`
	Package   string              `json:"package"`
	Version   string              `json:"version"`
	Classes   []SymbolClass       `json:"classes,omitempty"`
	Functions []SymbolCallable    `json:"functions,omitempty"`
	Commands  []CommandDefinition `json:"commands,omitempty"`
}

type CommandDefinition struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description" yaml:"description"`
	Usage       string `json:"usage,omitempty" yaml:"usage,omitempty"`
	Protected   bool   `json:"protected,omitempty" yaml:"protected,omitempty"`
	Handler     string `json:"handler,omitempty" yaml:"handler,omitempty"`
}

type SymbolClass struct {
	Name       string           `json:"name"`
	SuperClass string           `json:"super_class,omitempty"`
	Methods    []SymbolCallable `json:"methods,omitempty"`
	Properties []string         `json:"properties,omitempty"`
}

type SymbolCallable struct {
	Name       string            `json:"name"`
	Parameters []SymbolParameter `json:"parameters,omitempty"`
	ReturnType string            `json:"return_type,omitempty"`
}

type SymbolParameter struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

func BuildSymbolIndex(program *parser.Program, packageName, version string) SymbolIndex {
	index := SymbolIndex{Schema: SymbolSchemaVersion, Package: packageName, Version: version}
	if program == nil {
		return index
	}
	for _, statement := range program.Statements {
		switch node := statement.(type) {
		case *parser.ClassStatement:
			class := SymbolClass{Name: identifierValue(node.Name)}
			if node.SuperClass != nil {
				class.SuperClass = identifierValue(node.SuperClass)
			}
			if node.Body != nil {
				for _, member := range node.Body.Statements {
					switch value := member.(type) {
					case *parser.MethodStatement:
						class.Methods = append(class.Methods, callableSymbol(value.Name, value.Parameters, value.ReturnType))
					case *parser.LetStatement:
						if value.Name != nil {
							class.Properties = append(class.Properties, identifierValue(value.Name))
						}
					case *parser.MultiLetStatement:
						for _, declaration := range value.Declarations {
							class.Properties = append(class.Properties, identifierValue(declaration.Name))
						}
					}
				}
			}
			index.Classes = append(index.Classes, class)
		case *parser.MethodStatement:
			index.Functions = append(index.Functions, callableSymbol(node.Name, node.Parameters, node.ReturnType))
		}
	}
	return index
}

func callableSymbol(name *parser.Identifier, parameters []*parser.Parameter, returnType parser.Token) SymbolCallable {
	callable := SymbolCallable{Name: identifierValue(name), ReturnType: returnType.Literal}
	for _, parameter := range parameters {
		if parameter == nil || parameter.Name == nil {
			continue
		}
		typeName := parameter.Type.Literal
		if parameter.Type.Type == parser.VAR {
			typeName = ""
		}
		callable.Parameters = append(callable.Parameters, SymbolParameter{
			Name: identifierValue(parameter.Name),
			Type: typeName,
		})
	}
	return callable
}

func identifierValue(identifier *parser.Identifier) string {
	if identifier == nil {
		return ""
	}
	return identifier.Value
}

// BuildSymbolIndexFromCallables creates a SymbolIndex from pre-extracted classes and functions.
func BuildSymbolIndexFromCallables(packageName, version string, classes []SymbolClass, functions []SymbolCallable) SymbolIndex {
	return SymbolIndex{
		Schema:    SymbolSchemaVersion,
		Package:   packageName,
		Version:   version,
		Classes:   classes,
		Functions: functions,
	}
}
