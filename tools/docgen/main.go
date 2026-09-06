// docgen projects the executed native registry into documentation. It never
// invents parameter signatures from absent analyzer metadata.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
)

func main() {
	check := flag.Bool("check", false, "check the committed native documentation")
	flag.Parse()
	r := &core.Runtime{Variables: map[string]interface{}{}, Classes: map[string]*parser.ClassStatement{}, NativeHandlers: map[string]core.NativeHandler{}}
	r.RegisterNativeClasses()
	handlers := map[string]string{}
	locations := map[string]string{}
	fset := token.NewFileSet()
	paths, err := filepath.Glob("pkg/core/*.go")
	must(err)
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := goparser.ParseFile(fset, path, nil, 0)
		must(err)
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				locations[fn.Name.Name] = fmt.Sprintf("../%s#L%d", filepath.ToSlash(path), fset.Position(fn.Pos()).Line)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 3 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "registerNative" {
				return true
			}
			name, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				return true
			}
			class, err := strconv.Unquote(name.Value)
			must(err)
			if handler, ok := call.Args[2].(*ast.SelectorExpr); ok {
				handlers[class] = handler.Sel.Name
			}
			return true
		})
	}
	var output bytes.Buffer
	output.WriteString("# Catálogo nativo generado\n\n")
	output.WriteString("Antes: [biblioteca y contratos](MODULOS_NATIVOS.md). Índice: [documentación](README.md).\n\n")
	output.WriteString("Generado con `go run ./tools/docgen` desde `Runtime.RegisterNativeClasses()`.\n")
	output.WriteString("No editar manualmente. Cada nombre es una entrada del registro real. **Retorno publicado**\nno significa retorno exhaustivo observado: las discrepancias están en la biblioteca.\n")
	output.WriteString("Los parámetros nativos no están publicados en este registro; consulta las tablas\nde contratos y el handler enlazado. `mixed` no certifica aislamiento ni ausencia de fallos.\n\n")
	classes := make([]string, 0, len(r.Classes))
	for name := range r.Classes {
		classes = append(classes, name)
	}
	sort.Strings(classes)
	methodCount := 0
	for _, name := range classes {
		fmt.Fprintf(&output, "## %s\n\n", name)
		if loc, ok := locations[handlers[name]]; ok {
			fmt.Fprintf(&output, "Implementación: [%s](%s).\n\n", handlers[name], loc)
		}
		methods := map[string]string{}
		for _, stmt := range r.Classes[name].Body.Statements {
			if method, ok := stmt.(*parser.MethodStatement); ok {
				methods[method.Name.Value] = method.ReturnType.Literal
			}
		}
		if len(methods) == 0 {
			output.WriteString("Clase base registrada sin métodos nativos propios.\n\n")
			continue
		}
		names := make([]string, 0, len(methods))
		for method := range methods {
			names = append(names, method)
		}
		sort.Strings(names)
		output.WriteString("| Método | Retorno publicado al analizador |\n|---|---|\n")
		for _, method := range names {
			fmt.Fprintf(&output, "| `%s` | `%s` |\n", method, strings.ReplaceAll(methods[method], "|", "\\|"))
			methodCount++
		}
		output.WriteString("\n")
	}
	builtins := core.GetBuiltinFunctionNames()
	sort.Strings(builtins)
	output.WriteString("## Funciones globales\n\n")
	output.WriteString("Contratos: [funciones globales](FUNCIONES_GLOBALES.md). Variantes que comparten implementación\nse explican juntas allí; esta lista conserva todas las grafías registradas.\n\n")
	for _, name := range builtins {
		fmt.Fprintf(&output, "- `%s`\n", name)
	}
	fmt.Fprintf(&output, "\nTotal: %d clases, %d métodos y %d built-ins.\n", len(classes), methodCount, len(builtins))
	path := "docs/CATALOGO_NATIVO.md"
	if *check {
		disk, err := os.ReadFile(path)
		must(err)
		if !bytes.Equal(disk, output.Bytes()) {
			fmt.Fprintln(os.Stderr, "native documentation is stale: go run ./tools/docgen")
			os.Exit(1)
		}
	} else {
		must(os.WriteFile(path, output.Bytes(), 0644))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
