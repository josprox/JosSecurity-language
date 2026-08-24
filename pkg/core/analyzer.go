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

// Known framework singletons and native classes built into Joss
var frameworkClasses = map[string]bool{
	"GranDB": true, "Request": true, "Response": true, "Session": true,
	"Auth": true, "UserStorage": true, "SEO": true, "MFA": true,
	"TwoFactor": true, "UUID": true, "Http": true, "Markdown": true,
	"Schema": true, "Router": true, "System": true, "Str": true,
	"Math": true, "JSON": true, "Server": true, "View": true,
	"DB": true, "Cron": true, "Env": true, "Lang": true,
	"Config": true, "Turnstile": true, "Cache": true, "Log": true,
	"Event": true, "Storage": true, "Cookie": true, "Validator": true,
	"Mail": true, "Queue": true,
}

func isFrameworkClass(name string) bool {
	return frameworkClasses[name]
}

func isSpecialIdentifier(name string) bool {
	if name == "this" || name == "null" || name == "nil" || name == "default" {
		return true
	}
	// Constant style (ALL_CAPS) or environment variables
	if strings.ToUpper(name) == name && len(name) > 1 {
		return true
	}
	return false
}

// AnalyzeProgram performs static inspection of AST to detect unused/undeclared symbols.
func AnalyzeProgram(program *parser.Program) *AnalysisReport {
	report := &AnalysisReport{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	if program == nil {
		return report
	}

	declaredVars := make(map[string]int)    // varName -> line
	usedVars := make(map[string]bool)       // varName -> used
	declaredFuncs := make(map[string]int)   // funcName -> line
	calledFuncs := make(map[string]bool)    // funcName -> called
	declaredClasses := make(map[string]int) // className -> line
	usedClasses := make(map[string]bool)    // className -> used

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
			if isSpecialIdentifier(e.Value) || isFrameworkClass(e.Value) {
				return
			}
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
			if !isFrameworkClass(className) {
				if _, isDeclared := declaredClasses[className]; !isDeclared {
					report.Errors = append(report.Errors, fmt.Sprintf("Línea %d: Intento de instanciar la clase '%s' con 'new', pero la clase no está declarada.", e.Class.Token.Line, className))
				}
			}
			for _, arg := range e.Arguments {
				inspectExpression(arg)
			}

		case *parser.MemberExpression:
			if leftIdent, ok := e.Left.(*parser.Identifier); ok {
				name := leftIdent.Value
				if isFrameworkClass(name) || declaredClasses[name] > 0 || name == "this" || isSpecialIdentifier(name) {
					usedClasses[name] = true
				} else {
					inspectExpression(e.Left)
				}
			} else {
				inspectExpression(e.Left)
			}

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
					// Ignore 'default' keyword in match arm keys
					if ident, ok := k.(*parser.Identifier); ok && ident.Value == "default" {
						continue
					}
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
		if !usedVars[vName] && !isSpecialIdentifier(vName) {
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
		if isFrameworkOrMVCClass(cName) || usedClasses[cName] {
			continue
		}
		report.Warnings = append(report.Warnings, fmt.Sprintf("Línea %d: La clase '%s' está definida pero nunca es instanciada ni utilizada.", line, cName))
	}

	return report
}

func isFrameworkOrMVCClass(cName string) bool {
	if cName == "Main" || cName == "App" {
		return true
	}
	if strings.HasSuffix(cName, "Controller") ||
		strings.HasSuffix(cName, "Model") ||
		strings.HasSuffix(cName, "Middleware") ||
		strings.HasSuffix(cName, "Guard") ||
		strings.HasSuffix(cName, "Service") ||
		strings.HasSuffix(cName, "Downloader") ||
		strings.HasSuffix(cName, "Loader") ||
		strings.HasSuffix(cName, "Helper") {
		return true
	}
	if strings.HasPrefix(cName, "Create") ||
		strings.HasPrefix(cName, "Add") ||
		strings.HasPrefix(cName, "Drop") ||
		strings.HasPrefix(cName, "Seed") ||
		strings.HasPrefix(cName, "Update") ||
		strings.HasPrefix(cName, "Alter") {
		return true
	}
	// Standard ORM Model class names (e.g. Friendship, PubPackage, Project, CmsPost, etc.)
	if cName == "Friendship" || cName == "PubPackage" || cName == "Project" ||
		cName == "PubDownload" || cName == "OtpAccount" || cName == "Repository" ||
		cName == "Credential" || cName == "PubVersion" || cName == "Category" ||
		cName == "CmsPost" || cName == "User" {
		return true
	}
	return false
}
