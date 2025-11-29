package core

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/parser"
)

// Execute runs the parsed program
func (r *Runtime) Execute(program *parser.Program) {
	// Ensure env is loaded
	if len(r.Env) == 0 {
		r.LoadEnv()
	}

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
	// Execute imports first if they are at top level (outside class)
	for _, stmt := range program.Statements {
		if importStmt, ok := stmt.(*parser.ImportStatement); ok {
			r.executeImport(importStmt)
		}
	}

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
		// fmt.Println("Error: No se encontró la clase Main")
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

func (r *Runtime) executeBlock(block *parser.BlockStatement) interface{} {
	var result interface{}
	for _, stmt := range block.Statements {
		result = r.executeStatement(stmt)
	}
	return result
}

func (r *Runtime) registerClass(stmt *parser.ClassStatement) {
	r.Classes[stmt.Name.Value] = stmt
}

func (r *Runtime) executeStatement(stmt parser.Statement) interface{} {
	switch s := stmt.(type) {
	case *parser.LetStatement:
		val := r.evaluateExpression(s.Value)
		// Strict Typing: Store type
		r.VarTypes[s.Name.Value] = s.Token.Literal
		if !r.checkType(val, s.Token.Literal) {
			panic(fmt.Sprintf("Error de Tipado: Variable '%s' definida como '%s' pero asignada valor incompatible", s.Name.Value, s.Token.Literal))
		}
		r.Variables[s.Name.Value] = val
	case *parser.ExpressionStatement:
		return r.evaluateExpression(s.Expression)
	case *parser.ForeachStatement:
		return r.executeForeach(s)
	case *parser.ImportStatement:
		return r.executeImport(s)
	case *parser.EchoStatement:
		val := r.evaluateExpression(s.Value)
		fmt.Println(val)
	case *parser.WhileStatement:
		return r.executeWhile(s)
	case *parser.DoWhileStatement:
		return r.executeDoWhile(s)
	case *parser.TryCatchStatement:
		return r.executeTryCatch(s)
	case *parser.ThrowStatement:
		return r.executeThrow(s)
	case *parser.ReturnStatement:
		return r.executeReturn(s)
	case *parser.MethodStatement:
		r.Functions[s.Name.Value] = s
	case *parser.IfStatement:
		return r.executeIf(s)
	}
	return nil
}

func (r *Runtime) executeReturn(rs *parser.ReturnStatement) interface{} {
	if rs.ReturnValue != nil {
		return r.evaluateExpression(rs.ReturnValue)
	}
	return nil
}

func (r *Runtime) executeIf(is *parser.IfStatement) interface{} {
	cond := r.evaluateExpression(is.Condition)
	if isTruthy(cond) {
		return r.executeBlock(is.Consequence)
	} else if is.Alternative != nil {
		return r.executeBlock(is.Alternative)
	}
	return nil
}

func (r *Runtime) executeImport(stmt *parser.ImportStatement) interface{} {
	filename := stmt.Path

	// Handle Global Import
	if filename == "global" {
		filename = "config/global.joss"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// Try looking in parent directories if running from subfolder
			if _, err := os.Stat("../config/global.joss"); err == nil {
				filename = "../config/global.joss"
			} else if _, err := os.Stat("../../config/global.joss"); err == nil {
				filename = "../../config/global.joss"
			} else {
				fmt.Println("Error: @import \"global\" requiere 'config/global.joss'")
				return nil
			}
		}
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error: No se pudo importar '%s': %v\n", filename, err)
		return nil
	}

	l := parser.NewLexer(string(content))
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Printf("Error de parseo en '%s':\n", filename)
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return nil
	}

	// Execute imported program in current runtime (shared scope)
	for _, s := range program.Statements {
		r.executeStatement(s)
	}

	return nil
}

func (r *Runtime) executeForeach(fs *parser.ForeachStatement) interface{} {
	iterable := r.evaluateExpression(fs.Iterable)

	if list, ok := iterable.([]interface{}); ok {
		for _, item := range list {
			r.Variables[fs.Value] = item
			r.executeBlock(fs.Body)
		}
	} else if ch, ok := iterable.(*Channel); ok {
		for item := range ch.Ch {
			r.Variables[fs.Value] = item
			r.executeBlock(fs.Body)
		}
	} else {
		fmt.Printf("Error: Foreach espera un array o canal, se obtuvo: %T\n", iterable)
	}
	return nil
}

func (r *Runtime) executeWhile(ws *parser.WhileStatement) interface{} {
	for {
		cond := r.evaluateExpression(ws.Condition)
		if !isTruthy(cond) {
			break
		}
		r.executeBlock(ws.Body)
	}
	return nil
}

func (r *Runtime) executeDoWhile(dws *parser.DoWhileStatement) interface{} {
	for {
		r.executeBlock(dws.Body)
		cond := r.evaluateExpression(dws.Condition)
		if !isTruthy(cond) {
			break
		}
	}
	return nil
}

func (r *Runtime) executeTryCatch(tcs *parser.TryCatchStatement) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			// Catch the error
			// If err is a string (from throw "msg"), use it.
			// If it's a runtime panic, convert to string.
			var errVal interface{} = err
			if e, ok := err.(error); ok {
				errVal = e.Error()
			}

			// Bind error variable
			r.Variables[tcs.CatchVar] = errVal

			// Execute catch block
			result = r.executeBlock(tcs.CatchBlock)
		}
	}()

	return r.executeBlock(tcs.TryBlock)
}

func (r *Runtime) executeThrow(ts *parser.ThrowStatement) interface{} {
	val := r.evaluateExpression(ts.Value)
	panic(val)
	return nil
}
