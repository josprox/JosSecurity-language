package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

// RegisterNativeClasses injects the native class definitions into the runtime
func (r *Runtime) RegisterNativeClasses() {
	// Stack
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "Stack"},
		Body: &parser.BlockStatement{},
	})

	// Queue
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "Queue"},
		Body: &parser.BlockStatement{},
	})

	// GranMySQL
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "GranMySQL"},
		Body: &parser.BlockStatement{},
	})

	// Auth
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "Auth"},
		Body: &parser.BlockStatement{},
	})
}

func (r *Runtime) executeNativeMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch instance.Class.Name.Value {
	case "Stack":
		return r.executeStackMethod(instance, method, args)
	case "Queue":
		return r.executeQueueMethod(instance, method, args)
	case "GranMySQL":
		return r.executeGranMySQLMethod(instance, method, args)
	case "Auth":
		return r.executeAuthMethod(instance, method, args)
	}
	return nil
}

// Stack Implementation
func (r *Runtime) executeStackMethod(instance *Instance, method string, args []interface{}) interface{} {
	// ... (existing code)
	if _, ok := instance.Fields["_data"]; !ok {
		instance.Fields["_data"] = []interface{}{}
	}
	data := instance.Fields["_data"].([]interface{})

	switch method {
	case "push":
		if len(args) > 0 {
			instance.Fields["_data"] = append(data, args[0])
		}
		return nil
	case "pop":
		if len(data) == 0 {
			return nil
		}
		val := data[len(data)-1]
		instance.Fields["_data"] = data[:len(data)-1]
		return val
	case "peek":
		if len(data) == 0 {
			return nil
		}
		return data[len(data)-1]
	}
	return nil
}

// Queue Implementation
func (r *Runtime) executeQueueMethod(instance *Instance, method string, args []interface{}) interface{} {
	if _, ok := instance.Fields["_data"]; !ok {
		instance.Fields["_data"] = []interface{}{}
	}
	data := instance.Fields["_data"].([]interface{})

	switch method {
	case "enqueue":
		if len(args) > 0 {
			instance.Fields["_data"] = append(data, args[0])
		}
		return nil
	case "dequeue":
		if len(data) == 0 {
			return nil
		}
		val := data[0]
		instance.Fields["_data"] = data[1:]
		return val
	case "peek":
		if len(data) == 0 {
			return nil
		}
		return data[0]
	}
	return nil
}

// GranMySQL Implementation
func (r *Runtime) executeGranMySQLMethod(instance *Instance, method string, args []interface{}) interface{} {
	// ... (existing code)
	switch method {
	case "where":
		// Simulate query
		table := instance.Fields["tabla"]
		col := instance.Fields["comparar"]
		val := instance.Fields["comparable"]
		fmt.Printf("[GranMySQL] SELECT * FROM js_%v WHERE %v = %v\n", table, col, val)
		return "[{\"id\": 1, \"name\": \"Simulated User\"}]" // JSON Mock
	case "clasic":
		table := instance.Fields["tabla"]
		fmt.Printf("[GranMySQL] SELECT * FROM js_%v\n", table)
		return "[{\"id\": 1}, {\"id\": 2}]"
	}
	return nil
}

// Auth Implementation
func (r *Runtime) executeAuthMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "create":
		// Simulates Auth::create(["email" => "...", "password" => "..."])
		// Since we don't have maps yet, we accept a list or just print success.
		fmt.Println("[Security] Auth::create - Usuario registrado con Bcrypt (Simulado)")
		return true
	case "attempt":
		if len(args) >= 1 {
			email := args[0]
			fmt.Printf("[Security] Auth::attempt - Verificando credenciales para %v...\n", email)
			return "JOSS_SESSION_TOKEN_XYZ"
		}
	}
	return nil
}
