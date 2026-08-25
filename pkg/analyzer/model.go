// Package analyzer performs project-aware semantic analysis over the Joss AST.
// It deliberately depends only on language-layer packages; runtime symbols are
// supplied through Environment so the analyzer remains reusable and testable.
package analyzer

import (
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

type SourceUnit struct {
	Path    string
	Program *parser.Program
}

type Parameter struct {
	Name       string
	Type       typesystem.Type
	HasDefault bool
}

type Callable struct {
	Name       string
	Parameters []Parameter
	ReturnType typesystem.Type
	Variadic   bool
}

type Class struct {
	Name       string
	SuperClass string
	Methods    map[string]Callable
}

type Environment struct {
	Builtins map[string]Callable
	Classes  map[string]Class
	Globals  map[string]typesystem.Type
}

func NewEnvironment() Environment {
	return Environment{
		Builtins: make(map[string]Callable),
		Classes:  make(map[string]Class),
		Globals:  make(map[string]typesystem.Type),
	}
}
