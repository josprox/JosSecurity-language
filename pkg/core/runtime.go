package core

import (
	"fmt"
	"strconv"

	"github.com/jossecurity/joss/pkg/parser"
)

// Runtime manages the execution environment of a Joss program
type Runtime struct {
	Env       map[string]string
	Variables map[string]string
}

// NewRuntime creates a new Joss runtime
func NewRuntime() *Runtime {
	return &Runtime{
		Env:       make(map[string]string),
		Variables: make(map[string]string),
	}
}

// LoadEnv simulates loading and decrypting the environment
func (r *Runtime) LoadEnv() {
	fmt.Println("[Security] Cargando entorno encriptado en RAM...")
	// Simulation of decryption
	r.Env["APP_ENV"] = "dev"
	r.Env["PORT"] = "8000"
}

// Execute runs the parsed program
func (r *Runtime) Execute(program *parser.Program) {
	// First pass: Register classes and functions
	for _, stmt := range program.Statements {
		if classStmt, ok := stmt.(*parser.ClassStatement); ok {
			r.registerClass(classStmt)
		}
	}

	// Find and execute Main class Init main
	hasClasses := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*parser.ClassStatement); ok {
			hasClasses = true
			break
		}
	}

	if hasClasses {
		r.executeMain(program)
	} else {
		// Legacy mode (Phase 2 scripts)
		for _, stmt := range program.Statements {
			r.executeStatement(stmt)
		}
	}
}

func (r *Runtime) executeMain(program *parser.Program) {
	// Find Class Main
	var mainClass *parser.ClassStatement
	for _, stmt := range program.Statements {
		if s, ok := stmt.(*parser.ClassStatement); ok {
			if s.Name.Value == "Main" {
				mainClass = s
				break
			}
		}
	}

	if mainClass == nil {
		fmt.Println("Error: No se encontró la clase Main")
		return
	}

	// Find Init main inside Main
	var initMain *parser.InitStatement
	for _, stmt := range mainClass.Body.Statements {
		if s, ok := stmt.(*parser.InitStatement); ok {
			if s.Name.Value == "main" {
				initMain = s
				break
			}
		}
	}

	if initMain == nil {
		fmt.Println("Error: No se encontró Init main() en la clase Main")
		return
	}

	// Execute Init main body
	r.executeBlock(initMain.Body)
}

func (r *Runtime) executeBlock(block *parser.BlockStatement) string {
	var result string
	for _, stmt := range block.Statements {
		result = r.executeStatement(stmt)
	}
	return result
}

func (r *Runtime) registerClass(stmt *parser.ClassStatement) {
	// TODO: Store class definition
}

func (r *Runtime) executeStatement(stmt parser.Statement) string {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		val := r.evaluateExpression(s.Value)
		r.Variables[s.Name.Value] = val
	case *parser.ExpressionStatement:
		return r.evaluateExpression(s.Expression)
	}
	return ""
}

func (r *Runtime) evaluateExpression(exp parser.Expression) string {
	switch e := exp.(type) {
	case *parser.StringLiteral:
		return e.Value
	case *parser.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *parser.Boolean:
		return fmt.Sprintf("%t", e.Value)
	case *parser.CallExpression:
		return r.executeCall(e)
	case *parser.Identifier:
		if val, ok := r.Variables[e.Value]; ok {
			return val
		}
		return "undefined"
	case *parser.TernaryExpression:
		return r.evaluateTernary(e)
	case *parser.InfixExpression:
		return r.evaluateInfix(e)
	}
	return ""
}

func (r *Runtime) executeCall(call *parser.CallExpression) string {
	// Check if it's a print call
	if ident, ok := call.Function.(*parser.Identifier); ok {
		if ident.Value == "print" {
			for _, arg := range call.Arguments {
				val := r.evaluateExpression(arg)
				fmt.Println(val)
			}
		}
	}
	return ""
}

func (r *Runtime) evaluateTernary(te *parser.TernaryExpression) string {
	cond := r.evaluateExpression(te.Condition)
	isTrue := cond == "true" || cond == "TRUE"

	if isTrue {
		if te.True != nil {
			return r.executeBlock(te.True)
		}
	} else {
		if te.False != nil {
			return r.executeBlock(te.False)
		}
	}
	return ""
}

func (r *Runtime) evaluateInfix(ie *parser.InfixExpression) string {
	left := r.evaluateExpression(ie.Left)
	right := r.evaluateExpression(ie.Right)

	lInt, errL := strconv.ParseInt(left, 10, 64)
	rInt, errR := strconv.ParseInt(right, 10, 64)

	if errL == nil && errR == nil {
		switch ie.Operator {
		case "+":
			return fmt.Sprintf("%d", lInt+rInt)
		case "<":
			return fmt.Sprintf("%t", lInt < rInt)
		case ">":
			return fmt.Sprintf("%t", lInt > rInt)
		case "==":
			return fmt.Sprintf("%t", lInt == rInt)
		case "!=":
			return fmt.Sprintf("%t", lInt != rInt)
		case "<=":
			return fmt.Sprintf("%t", lInt <= rInt)
		case ">=":
			return fmt.Sprintf("%t", lInt >= rInt)
		}
	}

	return ""
}
