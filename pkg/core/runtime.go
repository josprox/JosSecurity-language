package core

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jossecurity/joss/pkg/parser"
)

// Runtime manages the execution environment of a Joss program
type Runtime struct {
	Env       map[string]string
	Variables map[string]interface{}
	Classes   map[string]*parser.ClassStatement
	DB        *sql.DB
}

type Instance struct {
	Class  *parser.ClassStatement
	Fields map[string]interface{}
}

type Cout struct{}
type Cin struct{}

// Future represents an asynchronous computation
type Future struct {
	done   chan bool
	result interface{}
	err    error
}

// Wait blocks until the Future completes and returns the result
func (f *Future) Wait() interface{} {
	<-f.done
	if f.err != nil {
		panic(f.err)
	}
	return f.result
}

func (c *Cout) String() string { return "cout" }
func (c *Cin) String() string  { return "cin" }

// NewRuntime creates a new Joss runtime
func NewRuntime() *Runtime {
	r := &Runtime{
		Env:       make(map[string]string),
		Variables: make(map[string]interface{}),
		Classes:   make(map[string]*parser.ClassStatement),
	}
	r.Variables["cout"] = &Cout{}
	r.Variables["cin"] = &Cin{}

	r.RegisterNativeClasses()

	return r
}

// LoadEnv simulates loading and decrypting the environment
func (r *Runtime) LoadEnv() {
	fmt.Println("[Security] Cargando entorno encriptado en RAM...")
	// Simulation of decryption
	r.Env["APP_ENV"] = "dev"
	r.Env["PORT"] = "8000"
	r.Env["PORT"] = "8000"
	r.Env["DB_HOST"] = "localhost"
	r.Env["DB_USER"] = "root"
	r.Env["DB_PASS"] = ""
	r.Env["DB_NAME"] = "prueba_joss"

	// Connect to DB
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", r.Env["DB_USER"], r.Env["DB_PASS"], r.Env["DB_HOST"], r.Env["DB_NAME"])
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("[GranMySQL] Error conectando a DB: %v\n", err)
	} else {
		// Test connection
		err = db.Ping()
		if err != nil {
			fmt.Printf("[GranMySQL] Error ping DB: %v\n", err)
		} else {
			fmt.Println("[GranMySQL] Conexión establecida con éxito")
			r.DB = db
		}
	}
}

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
	}
	return nil
}

func (r *Runtime) executeReturn(rs *parser.ReturnStatement) interface{} {
	if rs.ReturnValue != nil {
		return r.evaluateExpression(rs.ReturnValue)
	}
	return nil
}

func (r *Runtime) executeImport(stmt *parser.ImportStatement) interface{} {
	filename := stmt.Path
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
	} else {
		fmt.Printf("Error: Foreach espera un array, se obtuvo: %T\n", iterable)
	}
	return nil
}

func (r *Runtime) evaluateExpression(exp parser.Expression) interface{} {
	switch e := exp.(type) {
	case *parser.StringLiteral:
		return e.Value
	case *parser.IntegerLiteral:
		return e.Value
	case *parser.FloatLiteral:
		return e.Value
	case *parser.Boolean:
		return e.Value
	case *parser.CallExpression:
		return r.executeCall(e)
	case *parser.Identifier:
		if val, ok := r.Variables[e.Value]; ok {
			return val
		}
		return nil
	case *parser.TernaryExpression:
		return r.evaluateTernary(e)
	case *parser.InfixExpression:
		return r.evaluateInfix(e)
	case *parser.ArrayLiteral:
		return r.evaluateArray(e)
	case *parser.MapLiteral:
		return r.evaluateMap(e)
	case *parser.IndexExpression:
		return r.evaluateIndex(e)
	case *parser.NewExpression:
		return r.evaluateNew(e)
	case *parser.MemberExpression:
		return r.evaluateMember(e)
	case *parser.AssignExpression:
		return r.evaluateAssign(e)
	case *parser.IssetExpression:
		return r.evaluateIsset(e)
	case *parser.EmptyExpression:
		return r.evaluateEmpty(e)
	}
	return nil
}

func (r *Runtime) evaluateAssign(ae *parser.AssignExpression) interface{} {
	val := r.evaluateExpression(ae.Value)

	if ident, ok := ae.Left.(*parser.Identifier); ok {
		r.Variables[ident.Value] = val
		return val
	}

	if member, ok := ae.Left.(*parser.MemberExpression); ok {
		left := r.evaluateExpression(member.Left)
		if instance, ok := left.(*Instance); ok {
			instance.Fields[member.Property.Value] = val
			return val
		}
		fmt.Printf("Error: Asignación a miembro de no-instancia: %v\n", left)
		return nil
	}

	fmt.Printf("Error: Asignación inválida a %T\n", ae.Left)
	return nil
}

func (r *Runtime) evaluateArray(al *parser.ArrayLiteral) []interface{} {
	elements := []interface{}{}
	for _, el := range al.Elements {
		elements = append(elements, r.evaluateExpression(el))
	}
	return elements
}

func (r *Runtime) evaluateMap(ml *parser.MapLiteral) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range ml.Pairs {
		key := r.evaluateExpression(k)
		val := r.evaluateExpression(v)
		if keyStr, ok := key.(string); ok {
			m[keyStr] = val
		} else {
			fmt.Printf("Error: Clave de mapa inválida: %v (se espera string)\n", key)
		}
	}
	return m
}

func (r *Runtime) evaluateIndex(ie *parser.IndexExpression) interface{} {
	left := r.evaluateExpression(ie.Left)
	index := r.evaluateExpression(ie.Index)

	if list, ok := left.([]interface{}); ok {
		if idx, ok := index.(int64); ok {
			if idx >= 0 && idx < int64(len(list)) {
				return list[idx]
			}
			fmt.Println("Error: Índice fuera de rango")
		} else {
			fmt.Println("Error: El índice debe ser un entero")
		}
		return nil
	}

	if m, ok := left.(map[string]interface{}); ok {
		if key, ok := index.(string); ok {
			if val, exists := m[key]; exists {
				return val
			}
			return nil
		}
		fmt.Println("Error: El índice de un mapa debe ser string")
		return nil
	}

	fmt.Println("Error: No se puede indexar algo que no es un array o mapa")
	return nil
}

func (r *Runtime) evaluateTernary(te *parser.TernaryExpression) interface{} {
	cond := r.evaluateExpression(te.Condition)
	isTrue := false

	if b, ok := cond.(bool); ok {
		isTrue = b
	} else if s, ok := cond.(string); ok {
		isTrue = s == "true" || s == "TRUE"
	}

	if isTrue {
		if te.True != nil {
			return r.executeBlock(te.True)
		}
	} else {
		if te.False != nil {
			return r.executeBlock(te.False)
		}
	}
	return nil
}

func (r *Runtime) evaluateInfix(ie *parser.InfixExpression) interface{} {
	left := r.evaluateExpression(ie.Left)

	// Handle cin >> $var (Special case: Right is not evaluated as expression, but as l-value)
	if ie.Operator == ">>" {
		if _, ok := left.(*Cin); ok {
			if ident, ok := ie.Right.(*parser.Identifier); ok {
				var input string
				fmt.Scanln(&input)
				r.Variables[ident.Value] = input
				return left // Return cin for chaining?
			}
			fmt.Println("Error: cin >> requiere una variable")
			return nil
		}
	}

	right := r.evaluateExpression(ie.Right)

	// Handle cout << val
	if ie.Operator == "<<" {
		if _, ok := left.(*Cout); ok {
			fmt.Print(right)
			return left // Return cout for chaining
		}
	}

	// Smart Numerics: Auto-promote to float if needed
	toFloat := func(val interface{}) (float64, bool) {
		if i, ok := val.(int64); ok {
			return float64(i), true
		}
		if f, ok := val.(float64); ok {
			return f, true
		}
		return 0, false
	}

	lFloat, lIsNum := toFloat(left)
	rFloat, rIsNum := toFloat(right)

	if lIsNum && rIsNum {
		// If division, always float
		if ie.Operator == "/" {
			return lFloat / rFloat
		}

		// If any operand is float, result is float
		isFloatOp := false
		if _, ok := left.(float64); ok {
			isFloatOp = true
		}
		if _, ok := right.(float64); ok {
			isFloatOp = true
		}

		if isFloatOp {
			switch ie.Operator {
			case "+":
				return lFloat + rFloat
			case "-":
				return lFloat - rFloat
			case "*":
				return lFloat * rFloat
			case "<":
				return lFloat < rFloat
			case ">":
				return lFloat > rFloat
			case ">=":
				return lFloat >= rFloat
			case "<=":
				return lFloat <= rFloat
			case "==":
				return lFloat == rFloat
			case "!=":
				return lFloat != rFloat
			}
		} else {
			// Integer operations
			lInt := int64(lFloat)
			rInt := int64(rFloat)
			switch ie.Operator {
			case "+":
				return lInt + rInt
			case "-":
				return lInt - rInt
			case "*":
				return lInt * rInt
			case "<":
				return lInt < rInt
			case ">":
				return lInt > rInt
			case ">=":
				return lInt >= rInt
			case "<=":
				return lInt <= rInt
			case "==":
				return lInt == rInt
			case "!=":
				return lInt != rInt
			}
		}
	}

	lStr := fmt.Sprintf("%v", left)
	rStr := fmt.Sprintf("%v", right)
	if ie.Operator == "+" {
		return lStr + rStr
	}
	if ie.Operator == "==" {
		return lStr == rStr
	}
	if ie.Operator == "!=" {
		return lStr != rStr
	}

	return nil
}

type BoundMethod struct {
	Method   *parser.MethodStatement
	Instance *Instance
}

func (r *Runtime) evaluateNew(ne *parser.NewExpression) interface{} {
	className := ne.Class.Value
	classStmt, ok := r.Classes[className]
	if !ok {
		fmt.Printf("Error: Clase '%s' no encontrada\n", className)
		return nil
	}

	instance := &Instance{
		Class:  classStmt,
		Fields: make(map[string]interface{}),
	}

	// Collect inheritance chain
	chain := []*parser.ClassStatement{classStmt}
	curr := classStmt
	for curr.SuperClass != nil {
		parentName := curr.SuperClass.Value
		if parent, ok := r.Classes[parentName]; ok {
			chain = append(chain, parent)
			curr = parent
		} else {
			break
		}
	}

	// Initialize properties (Parent -> Child)
	for i := len(chain) - 1; i >= 0; i-- {
		cls := chain[i]
		for _, stmt := range cls.Body.Statements {
			if let, ok := stmt.(*parser.LetStatement); ok {
				instance.Fields[let.Name.Value] = r.evaluateExpression(let.Value)
			}
		}
	}

	// Call constructor if exists
	for _, stmt := range classStmt.Body.Statements {
		if method, ok := stmt.(*parser.MethodStatement); ok {
			if method.Name.Value == "constructor" {
				r.callMethod(method, instance, ne.Arguments)
				break
			}
		}
		if initStmt, ok := stmt.(*parser.InitStatement); ok {
			if initStmt.Name.Value == "constructor" {
				// Convert to MethodStatement
				method := &parser.MethodStatement{
					Token:      initStmt.Token,
					Name:       initStmt.Name,
					Parameters: initStmt.Parameters,
					Body:       initStmt.Body,
				}
				r.callMethod(method, instance, ne.Arguments)
				break
			}
		}
	}

	return instance
}

func (r *Runtime) evaluateMember(me *parser.MemberExpression) interface{} {
	left := r.evaluateExpression(me.Left)
	instance, ok := left.(*Instance)
	if !ok {
		fmt.Printf("Error: %v no es una instancia\n", left)
		return nil
	}

	propName := me.Property.Value

	// Check fields
	if val, ok := instance.Fields[propName]; ok {
		return val
	}

	// Check methods (Function and Init)
	// Check methods (Function and Init)
	currentClass := instance.Class
	for currentClass != nil {
		for _, stmt := range currentClass.Body.Statements {
			if method, ok := stmt.(*parser.MethodStatement); ok {
				if method.Name.Value == propName {
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
			if initStmt, ok := stmt.(*parser.InitStatement); ok {
				if initStmt.Name.Value == propName {
					// Convert InitStatement to MethodStatement for compatibility
					method := &parser.MethodStatement{
						Token:      initStmt.Token,
						Name:       initStmt.Name,
						Parameters: initStmt.Parameters,
						Body:       initStmt.Body,
					}
					return &BoundMethod{Method: method, Instance: instance}
				}
			}
		}

		// Move to parent
		if currentClass.SuperClass != nil {
			parentName := currentClass.SuperClass.Value
			if parent, ok := r.Classes[parentName]; ok {
				currentClass = parent
			} else {
				// fmt.Printf("Error: Clase padre '%s' no encontrada\n", parentName)
				currentClass = nil
			}
		} else {
			currentClass = nil
		}
	}

	// Check for Native Class methods (Stack, Queue, GranMySQL, Auth)
	className := instance.Class.Name.Value
	if className == "Stack" || className == "Queue" || className == "GranMySQL" || className == "Auth" {
		// Return a synthetic method with nil body to trigger native execution
		return &BoundMethod{
			Method: &parser.MethodStatement{
				Name: &parser.Identifier{Value: propName},
				Body: nil,
			},
			Instance: instance,
		}
	}

	fmt.Printf("Error: Propiedad o método '%s' no encontrado\n", propName)
	return nil
}

func (r *Runtime) evaluateIsset(ie *parser.IssetExpression) bool {
	for _, arg := range ie.Arguments {
		if !r.checkExistence(arg) {
			return false
		}
	}
	return true
}

func (r *Runtime) checkExistence(exp parser.Expression) bool {
	switch e := exp.(type) {
	case *parser.Identifier:
		_, ok := r.Variables[e.Value]
		return ok
	case *parser.IndexExpression:
		left := r.evaluateExpression(e.Left)
		if list, ok := left.([]interface{}); ok {
			index := r.evaluateExpression(e.Index)
			if idx, ok := index.(int64); ok {
				return idx >= 0 && idx < int64(len(list))
			}
		}
		return false
	case *parser.MemberExpression:
		left := r.evaluateExpression(e.Left)
		if instance, ok := left.(*Instance); ok {
			_, ok := instance.Fields[e.Property.Value]
			return ok
		}
		return false
	}
	return false
}

func (r *Runtime) evaluateEmpty(ee *parser.EmptyExpression) bool {
	// Special case: if argument is variable/index/member that doesn't exist, return true
	if !r.checkExistence(ee.Argument) {
		return true
	}

	val := r.evaluateExpression(ee.Argument)
	return isFalsy(val)
}

func isFalsy(val interface{}) bool {
	if val == nil {
		return true
	}
	if b, ok := val.(bool); ok {
		return !b
	}
	if s, ok := val.(string); ok {
		return s == "" || s == "0"
	}
	if i, ok := val.(int64); ok {
		return i == 0
	}
	if list, ok := val.([]interface{}); ok {
		return len(list) == 0
	}
	return false
}

func (r *Runtime) callMethod(method *parser.MethodStatement, instance *Instance, args []parser.Expression) interface{} {
	// Native Method Support
	if method.Body == nil {
		evalArgs := []interface{}{}
		for _, arg := range args {
			evalArgs = append(evalArgs, r.evaluateExpression(arg))
		}
		return r.executeNativeMethod(instance, method.Name.Value, evalArgs)
	}

	// Save previous "this" if exists (for nested calls)
	prevThis := r.Variables["this"]
	r.Variables["this"] = instance

	// Bind arguments
	for i, param := range method.Parameters {
		if i < len(args) {
			val := r.evaluateExpression(args[i])
			r.Variables[param.Value] = val
		}
	}

	// Execute body
	res := r.executeBlock(method.Body)

	// Restore "this"
	if prevThis != nil {
		r.Variables["this"] = prevThis
	} else {
		delete(r.Variables, "this")
	}

	return res
}

func (r *Runtime) executeCall(call *parser.CallExpression) interface{} {
	fn := r.evaluateExpression(call.Function)

	if bound, ok := fn.(*BoundMethod); ok {
		return r.callMethod(bound.Method, bound.Instance, call.Arguments)
	}

	if ident, ok := call.Function.(*parser.Identifier); ok {
		if ident.Value == "print" || ident.Value == "echo" {
			for _, arg := range call.Arguments {
				val := r.evaluateExpression(arg)
				fmt.Println(val)
			}
			return nil
		}
		if ident.Value == "printf" {
			if len(call.Arguments) > 0 {
				format := r.evaluateExpression(call.Arguments[0])
				args := []interface{}{}
				for _, arg := range call.Arguments[1:] {
					args = append(args, r.evaluateExpression(arg))
				}
				if fmtStr, ok := format.(string); ok {
					fmt.Printf(fmtStr, args...)
				}
			}
			return nil
		}
		if ident.Value == "env" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if key, ok := arg.(string); ok {
					if val, exists := r.Env[key]; exists {
						return val
					}
					return ""
				}
			}
			return nil
		}
		if ident.Value == "len" || ident.Value == "count" {
			if len(call.Arguments) == 1 {
				arg := r.evaluateExpression(call.Arguments[0])
				if list, ok := arg.([]interface{}); ok {
					return int64(len(list))
				}
				if str, ok := arg.(string); ok {
					return int64(len(str))
				}
			}
			return int64(0)
		}
		// async function - executes code in goroutine
		if ident.Value == "async" {
			if len(call.Arguments) == 1 {
				future := &Future{
					done: make(chan bool),
				}

				// Execute the argument in a goroutine
				go func() {
					defer func() {
						if r := recover(); r != nil {
							future.err = fmt.Errorf("%v", r)
						}
						close(future.done)
					}()

					// Evaluate the argument
					future.result = r.evaluateExpression(call.Arguments[0])
				}()

				return future
			}
			return nil
		}
		// await function - waits for a Future
		if ident.Value == "await" {
			if len(call.Arguments) == 1 {
				futureVal := r.evaluateExpression(call.Arguments[0])
				if future, ok := futureVal.(*Future); ok {
					return future.Wait()
				}
				fmt.Println("Error: await expects a Future")
			}
			return nil
		}
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

func isTruthy(val interface{}) bool {
	return !isFalsy(val)
}
