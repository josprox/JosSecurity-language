package core

import (
	"fmt"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
)

// AnalysisReport holds errors and warnings discovered during static analysis.
type AnalysisReport struct {
	Errors   []string
	Warnings []string
}

// HasIssues returns true if any warnings or errors were detected.
func (ar *AnalysisReport) HasIssues() bool {
	return len(ar.Errors) > 0 || len(ar.Warnings) > 0
}

// PrintReport outputs the formatted analysis report to stdout.
func (ar *AnalysisReport) PrintReport() {
	if !ar.HasIssues() {
		fmt.Println("🔍 [MODO DEBUG] Análisis estático: Todo en orden. No se detectaron problemas de variables, funciones ni clases.")
		return
	}

	fmt.Println("\n🔍 [MODO DEBUG - ANÁLISIS ESTÁTICO CÓDIGO JOSS]")
	fmt.Println(strings.Repeat("-", 60))
	for _, w := range ar.Warnings {
		fmt.Printf("⚠️  [ADVERTENCIA] %s\n", w)
	}
	for _, e := range ar.Errors {
		fmt.Printf("❌ [ERROR] %s\n", e)
	}
	fmt.Println(strings.Repeat("-", 60))
}

// Known Joss built-in functions check using centralized core.IsBuiltin registry.

// AnalyzeProgram performs static inspection of AST to detect unused/undeclared symbols.
func AnalyzeProgram(program *parser.Program) *AnalysisReport {
	report := &AnalysisReport{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	if program == nil {
		return report
	}

	declaredVars := make(map[string]int)     // varName -> line
	usedVars := make(map[string]bool)        // varName -> used
	declaredFuncs := make(map[string]int)    // funcName -> line
	calledFuncs := make(map[string]bool)     // funcName -> called
	declaredClasses := make(map[string]int)  // className -> line
	usedClasses := make(map[string]bool)     // className -> used

	// Pass 1: Collect top-level declarations (functions, classes, global vars)
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.ClassStatement:
			declaredClasses[s.Name.Value] = s.Name.Token.Line
		case *parser.MethodStatement:
			declaredFuncs[s.Name.Value] = s.Name.Token.Line
		}
	}

	// Pass 2: Traverse AST statements to record definitions, usages, and calls
	var inspectStatement func(stmt parser.Statement)
	var inspectExpression func(exp parser.Expression)

	inspectExpression = func(exp parser.Expression) {
		if exp == nil {
			return
		}
		switch e := exp.(type) {
		case *parser.Identifier:
			usedVars[e.Value] = true
			// If not declared and not a known class/func, report undeclared var error
			if _, isDecl := declaredVars[e.Value]; !isDecl {
				if _, isClass := declaredClasses[e.Value]; !isClass {
					if _, isFunc := declaredFuncs[e.Value]; !isFunc {
						report.Errors = append(report.Errors, fmt.Sprintf("Línea %d: La variable '$%s' está siendo utilizada sin haber sido declarada previa ni formalmente.", e.Token.Line, e.Value))
					}
				}
			}

		case *parser.CallExpression:
			if ident, ok := e.Function.(*parser.Identifier); ok {
				fnName := ident.Value
				calledFuncs[fnName] = true
				if !IsBuiltin(fnName) {
					if _, isDeclared := declaredFuncs[fnName]; !isDeclared {
						report.Errors = append(report.Errors, fmt.Sprintf("Línea %d: La función '%s()' está siendo invocada pero no ha sido declarada.", ident.Token.Line, fnName))
					}
				}
			} else {
				inspectExpression(e.Function)
			}
			for _, arg := range e.Arguments {
				inspectExpression(arg)
			}

		case *parser.NewExpression:
			className := e.Class.Value
			usedClasses[className] = true
			if _, isDeclared := declaredClasses[className]; !isDeclared {
				report.Errors = append(report.Errors, fmt.Sprintf("Línea %d: Intento de instanciar la clase '%s' con 'new', pero la clase no está declarada.", e.Class.Token.Line, className))
			}
			for _, arg := range e.Arguments {
				inspectExpression(arg)
			}

		case *parser.MemberExpression:
			inspectExpression(e.Left)

		case *parser.InfixExpression:
			inspectExpression(e.Left)
			inspectExpression(e.Right)

		case *parser.PrefixExpression:
			inspectExpression(e.Right)

		case *parser.PostfixExpression:
			inspectExpression(e.Left)

		case *parser.TernaryExpression:
			inspectExpression(e.Condition)
			inspectExpression(e.True)
			// e.False handled if AST supports it

		case *parser.ArrayLiteral:
			for _, el := range e.Elements {
				inspectExpression(el)
			}

		case *parser.MapLiteral:
			for _, val := range e.Pairs {
				inspectExpression(val)
			}

		case *parser.AssignExpression:
			if ident, ok := e.Left.(*parser.Identifier); ok {
				// Mark as declared/initialized
				if _, ok := declaredVars[ident.Value]; !ok {
					declaredVars[ident.Value] = ident.Token.Line
				}
			} else {
				inspectExpression(e.Left)
			}
			inspectExpression(e.Value)

		case *parser.MatchExpression:
			inspectExpression(e.Subject)
			for _, arm := range e.Arms {
				for _, k := range arm.Keys {
					inspectExpression(k)
				}
				inspectExpression(arm.Value)
			}
		}
	}

	inspectStatement = func(stmt parser.Statement) {
		if stmt == nil {
			return
		}
		switch s := stmt.(type) {
		case *parser.LetStatement:
			declaredVars[s.Name.Value] = s.Name.Token.Line
			if s.Value != nil {
				inspectExpression(s.Value)
			}

		case *parser.MultiLetStatement:
			for _, decl := range s.Declarations {
				declaredVars[decl.Name.Value] = s.TypeToken.Line
				if decl.Value != nil {
					inspectExpression(decl.Value)
				}
			}

		case *parser.ExpressionStatement:
			inspectExpression(s.Expression)

		case *parser.EchoStatement:
			inspectExpression(s.Value)

		case *parser.ReturnStatement:
			if s.ReturnValue != nil {
				inspectExpression(s.ReturnValue)
			}

		case *parser.ThrowStatement:
			inspectExpression(s.Value)

		case *parser.WhileStatement:
			inspectExpression(s.Condition)
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.DoWhileStatement:
			inspectExpression(s.Condition)
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.ForeachStatement:
			inspectExpression(s.Iterable)
			declaredVars[s.Value] = s.Token.Line
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.TryCatchStatement:
			if s.TryBlock != nil {
				for _, bStmt := range s.TryBlock.Statements {
					inspectStatement(bStmt)
				}
			}
			if s.CatchVar != "" {
				declaredVars[s.CatchVar] = s.Token.Line
			}
			if s.CatchBlock != nil {
				for _, bStmt := range s.CatchBlock.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.ClassStatement:
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.MethodStatement:
			for _, param := range s.Parameters {
				if param.Name != nil {
					declaredVars[param.Name.Value] = param.Name.Token.Line
				}
			}
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}

		case *parser.InitStatement:
			if s.Body != nil {
				for _, bStmt := range s.Body.Statements {
					inspectStatement(bStmt)
				}
			}
		}
	}

	for _, stmt := range program.Statements {
		inspectStatement(stmt)
	}

	// Pass 3: Check for unused variables
	for vName, line := range declaredVars {
		if !usedVars[vName] {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Línea %d: La variable '$%s' fue declarada pero nunca se utiliza en el código.", line, vName))
		}
	}

	// Pass 4: Check for unused functions
	for fName, line := range declaredFuncs {
		if fName != "main" && fName != "Init" && !calledFuncs[fName] {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Línea %d: La función/método '%s()' está definida pero nunca es invocada.", line, fName))
		}
	}

	// Pass 5: Check for unused classes
	for cName, line := range declaredClasses {
		if cName != "Main" && !usedClasses[cName] {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Línea %d: La clase '%s' está definida pero nunca es instanciada ni utilizada.", line, cName))
		}
	}

	return report
}
