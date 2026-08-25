package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

// Execute runs the parsed program
func (r *Runtime) Execute(program *parser.Program) {
	// Ensure env is loaded
	if len(r.Env) == 0 {
		r.LoadEnv(nil)
	}

	// First pass: Register classes and functions
	for _, stmt := range program.Statements {
		if classStmt, ok := stmt.(*parser.ClassStatement); ok {
			r.registerClass(classStmt)
		}
		if methodStmt, ok := stmt.(*parser.MethodStatement); ok {
			r.Functions[methodStmt.Name.Value] = methodStmt
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
		hasMain := false
		for _, stmt := range program.Statements {
			if s, ok := stmt.(*parser.ClassStatement); ok && s.Name.Value == "Main" {
				hasMain = true
				break
			}
		}
		if hasMain {
			r.executeMain(program)
		} else {
			for _, stmt := range program.Statements {
				if _, ok := stmt.(*parser.ClassStatement); !ok {
					r.executeStatement(stmt)
				}
			}
		}
	} else {
		// Script programs execute their top-level statements directly.
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
		if r.Constants == nil {
			r.Constants = make(map[string]bool)
		}
		if r.Constants[s.Name.Value] {
			panic(&JossError{Type: "ConstantAssignment", Message: fmt.Sprintf("La constante '%s' no puede redeclararse", s.Name.Value), File: r.CurrentFile, Line: s.Name.Token.Line})
		}
		var val interface{}
		if s.Value != nil {
			val = r.evaluateExpression(s.Value)
			if s.Token.Literal != "var" {
				val = r.coerceToTypedValue(val, s.Token.Literal)
			}
		} else {
			val = r.getZeroValue(s.Token.Literal)
		}

		declaredType := s.Token.Literal
		if declaredType == "var" {
			declaredType = runtimeTypeName(val)
		}
		if declaredType != "" {
			r.VarTypes[s.Name.Value] = declaredType
		}
		if declaredType != "" && !r.checkType(val, declaredType) {
			panic(fmt.Sprintf("Error de Tipado: Variable '%s' definida como '%s' pero asignada valor incompatible", s.Name.Value, s.Token.Literal))
		}
		r.Variables[s.Name.Value] = val
		if s.IsConst {
			r.Constants[s.Name.Value] = true
		}
	case *parser.MultiLetStatement:
		// int $a,$b  or  int $a=1,$b=2
		for _, decl := range s.Declarations {
			var val interface{}
			if decl.Value != nil {
				val = r.evaluateExpression(decl.Value)
				if s.TypeToken.Literal != "var" {
					val = r.coerceToTypedValue(val, s.TypeToken.Literal)
				}
			} else {
				val = r.getZeroValue(s.TypeToken.Literal)
			}
			declaredType := s.TypeToken.Literal
			if declaredType == "var" {
				declaredType = runtimeTypeName(val)
			}
			if declaredType != "" {
				r.VarTypes[decl.Name.Value] = declaredType
			}
			if declaredType != "" && !r.checkType(val, declaredType) {
				panic(fmt.Sprintf("Error de Tipado: Variable '%s' definida como '%s' pero asignada valor incompatible", decl.Name.Value, s.TypeToken.Literal))
			}
			r.Variables[decl.Name.Value] = val
		}
	case *parser.ExpressionStatement:
		return r.evaluateExpression(s.Expression)
	case *parser.ForeachStatement:
		return r.executeForeach(s)
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
	case *parser.BreakStatement:
		return r.executeBreak(s)
	case *parser.ContinueStatement:
		return r.executeContinue(s)
	case *parser.MethodStatement:
		r.Functions[s.Name.Value] = s
	case *parser.ClassStatement:
		r.registerClass(s)

	}
	return nil
}

func (r *Runtime) executeReturn(rs *parser.ReturnStatement) interface{} {
	var val interface{}
	if rs.ReturnValue != nil {
		val = r.evaluateExpression(rs.ReturnValue)
	}
	panic(&ReturnPanic{Value: val})
}

func (r *Runtime) executeBreak(bs *parser.BreakStatement) interface{} {
	panic(&BreakPanic{})
}

func (r *Runtime) executeContinue(cs *parser.ContinueStatement) interface{} {
	panic(&ContinuePanic{})
}

func (r *Runtime) executeForeach(fs *parser.ForeachStatement) interface{} {
	iterable := r.evaluateExpression(fs.Iterable)

	executeIter := func(item interface{}) (shouldBreak bool) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case *BreakPanic:
					shouldBreak = true
				case *ContinuePanic:
					// Just return from this closure, which continues the loop
				default:
					panic(err) // Bubble up Returns and others
				}
			}
		}()
		r.Variables[fs.Value] = item
		r.executeBlock(fs.Body)
		return false
	}

	if list, ok := iterable.([]interface{}); ok {
		for _, item := range list {
			if executeIter(item) {
				break
			}
		}
	} else if list, ok := iterable.([]map[string]interface{}); ok {
		for _, item := range list {
			if executeIter(item) {
				break
			}
		}
	} else if ch, ok := iterable.(*Channel); ok {
		for item := range ch.Ch {
			if executeIter(item) {
				break
			}
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

		shouldBreak := false
		func() {
			defer func() {
				if err := recover(); err != nil {
					switch err.(type) {
					case *BreakPanic:
						shouldBreak = true
					case *ContinuePanic:
						// Skip
					default:
						panic(err)
					}
				}
			}()
			r.executeBlock(ws.Body)
		}()

		if shouldBreak {
			break
		}
	}
	return nil
}

func (r *Runtime) executeDoWhile(dws *parser.DoWhileStatement) interface{} {
	for {
		shouldBreak := false
		func() {
			defer func() {
				if err := recover(); err != nil {
					switch err.(type) {
					case *BreakPanic:
						shouldBreak = true
					case *ContinuePanic:
						// Skip
					default:
						panic(err)
					}
				}
			}()
			r.executeBlock(dws.Body)
		}()

		if shouldBreak {
			break
		}

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
			// Do NOT catch internal control flow panics
			switch err.(type) {
			case *ReturnPanic, *BreakPanic, *ContinuePanic:
				panic(err) // Let it bubble up
			}

			// Build the value exposed to the catch variable.
			// If it is a JossError, expose a map so Joss code can inspect fields.
			var errVal interface{}
			if je, ok := err.(*JossError); ok {
				errVal = map[string]interface{}{
					"message": je.Message,
					"type":    je.Type,
					"file":    je.File,
					"line":    int64(je.Line),
					"error":   je.Error(),
				}
			} else if e, ok := err.(error); ok {
				errVal = e.Error()
			} else {
				errVal = fmt.Sprintf("%v", err)
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
}
