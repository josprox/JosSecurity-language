package analyzer

import (
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/typesystem"
)

type symbolKind string

const (
	symbolVariable  symbolKind = "variable"
	symbolParameter symbolKind = "parameter"
	symbolIteration symbolKind = "iteration"
	symbolCatch     symbolKind = "catch"
	symbolImplicit  symbolKind = "implicit"
)

type symbol struct {
	Name      string
	Type      typesystem.Type
	Kind      symbolKind
	Token     parser.Token
	File      string
	Used      bool
	Dynamic   bool
	Inferred  bool
	Constant  bool
	Synthetic bool
}

type scope struct {
	parent  *scope
	symbols map[string]*symbol
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, symbols: make(map[string]*symbol)}
}

func cleanName(name string) string { return strings.TrimPrefix(name, "$") }

func (s *scope) local(name string) (*symbol, bool) {
	v, ok := s.symbols[cleanName(name)]
	return v, ok
}

func (s *scope) resolve(name string) (*symbol, bool) {
	name = cleanName(name)
	for current := s; current != nil; current = current.parent {
		if value, ok := current.symbols[name]; ok {
			return value, true
		}
	}
	return nil, false
}

func (s *scope) put(value *symbol) { s.symbols[cleanName(value.Name)] = value }
