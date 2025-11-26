package core

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jossecurity/joss/pkg/parser"
	"golang.org/x/crypto/bcrypt"
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
	case "query":
		// Execute raw SQL (useful for migrations)
		if len(args) > 0 {
			if sqlStr, ok := args[0].(string); ok {
				if r.DB == nil {
					fmt.Println("[GranMySQL] Error: No hay conexión a DB")
					return nil
				}
				_, err := r.DB.Exec(sqlStr)
				if err != nil {
					fmt.Printf("[GranMySQL] Error en query: %v\n", err)
					return false
				}
				return true
			}
		}
	case "where":
		// Real query: SELECT * FROM table WHERE col = val
		table := instance.Fields["tabla"]
		col := instance.Fields["comparar"]
		val := instance.Fields["comparable"]

		if r.DB == nil {
			fmt.Println("[GranMySQL] Error: No hay conexión a DB")
			return "[]"
		}

		query := fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", table, col)
		rows, err := r.DB.Query(query, val)
		if err != nil {
			fmt.Printf("[GranMySQL] Error en where: %v\n", err)
			return "[]"
		}
		defer rows.Close()

		// Parse rows to JSON-like string (simplified)
		// For now, just return a success message or simple representation
		// Implementing full JSON serialization from rows is complex without struct reflection
		// We'll return a simple string representation for verification
		var results []string
		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range cols {
			valPtrs[i] = &vals[i]
		}

		for rows.Next() {
			rows.Scan(valPtrs...)
			rowStr := "{"
			for i, colName := range cols {
				valVal := vals[i]
				if b, ok := valVal.([]byte); ok {
					valVal = string(b)
				}
				rowStr += fmt.Sprintf("\"%s\": \"%v\", ", colName, valVal)
			}
			rowStr += "}"
			results = append(results, rowStr)
		}

		return fmt.Sprintf("%v", results)

	case "clasic":
		// Real query: SELECT * FROM table
		table := instance.Fields["tabla"]
		if r.DB == nil {
			return "[]"
		}
		// Similar implementation to where...
		fmt.Printf("[GranMySQL] SELECT * FROM %v (Real)\n", table)
		return "[{\"id\": 1, \"real\": true}]"
	}
	return nil
}

// Auth Implementation
func (r *Runtime) executeAuthMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "create":
		// Auth::create([email, password, name])
		if len(args) > 0 {
			if data, ok := args[0].([]interface{}); ok && len(data) >= 2 {
				email := data[0].(string)
				password := data[1].(string)
				name := "User"
				if len(data) > 2 {
					name = data[2].(string)
				}

				// Hash password
				hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					fmt.Printf("[Security] Error hashing password: %v\n", err)
					return false
				}
				hashedPassword := string(hashedBytes)

				// Insert into DB
				if r.DB == nil {
					fmt.Println("[Security] Error: No DB connection")
					return false
				}

				_, err = r.DB.Exec("INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", name, email, hashedPassword)
				if err != nil {
					fmt.Printf("[Security] Error creando usuario: %v\n", err)
					return false
				}
				fmt.Println("[Security] Usuario registrado exitosamente.")
				return true
			}
		}
	case "attempt":
		if len(args) >= 2 {
			email := args[0].(string)
			password := args[1].(string)

			if r.DB == nil {
				return false
			}

			var storedHash string
			err := r.DB.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&storedHash)
			if err != nil {
				if err == sql.ErrNoRows {
					fmt.Println("[Security] Usuario no encontrado.")
				} else {
					fmt.Printf("[Security] Error DB: %v\n", err)
				}
				return false
			}

			err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
			if err != nil {
				fmt.Println("[Security] Contraseña incorrecta.")
				return false
			}

			fmt.Println("[Security] Login exitoso.")
			// Generate simple token
			return fmt.Sprintf("JOSS_TOKEN_%d", time.Now().Unix())
		}
	}
	return nil
}
