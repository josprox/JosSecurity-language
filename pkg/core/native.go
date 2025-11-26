package core

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	authClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Auth"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(authClass)
	r.Variables["Auth"] = &Instance{Class: authClass, Fields: make(map[string]interface{})}
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

// Helper to get table prefix
func getTablePrefix() string {
	prefix := os.Getenv("DB_PREFIX")
	if prefix == "" {
		return "js_"
	}
	return prefix
}

// GranMySQL Implementation
func (r *Runtime) executeGranMySQLMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Initialize internal state if needed
	if _, ok := instance.Fields["_wheres"]; !ok {
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_select"] = "*"
		instance.Fields["_table"] = ""
	}

	switch method {
	case "table":
		if len(args) > 0 {
			tableName := args[0].(string)
			// Apply prefix if not already present (and not a raw query)
			prefix := getTablePrefix()
			if !strings.HasPrefix(tableName, prefix) {
				tableName = prefix + tableName
			}
			instance.Fields["_table"] = tableName
		}
		return instance // Return this for chaining

	case "select":
		if len(args) > 0 {
			if cols, ok := args[0].(string); ok {
				instance.Fields["_select"] = cols
			} else if cols, ok := args[0].([]interface{}); ok {
				// Handle array of columns
				strCols := []string{}
				for _, c := range cols {
					strCols = append(strCols, fmt.Sprintf("%v", c))
				}
				instance.Fields["_select"] = strings.Join(strCols, ", ")
			}
		}
		return instance

	case "where":
		// Support both old and new API
		// Old API: where("json") - uses tabla, comparar, comparable properties
		// New API: where(col, val) or where(col, op, val) - fluent builder

		if len(args) == 1 {
			// Old API: where("json") or where("array")
			format := args[0].(string)

			// Use legacy properties
			table := instance.Fields["tabla"]
			col := instance.Fields["comparar"]
			val := instance.Fields["comparable"]

			if r.DB == nil {
				return "[]"
			}

			query := fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", table, col)
			rows, err := r.DB.Query(query, val)
			if err != nil {
				fmt.Printf("[GranMySQL] Error en where: %v\n", err)
				return "[]"
			}
			defer rows.Close()

			// Return based on format
			if format == "json" {
				return rowsToJSON(rows)
			}
			return rowsToJSON(rows) // Default to JSON
		}

		// New fluent builder API
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		if len(args) == 2 {
			col := args[0].(string)
			val := args[1]
			wheres = append(wheres, fmt.Sprintf("%s = ?", col))
			bindings = append(bindings, val)
		} else if len(args) == 3 {
			col := args[0].(string)
			op := args[1].(string)
			val := args[2]
			wheres = append(wheres, fmt.Sprintf("%s %s ?", col, op))
			bindings = append(bindings, val)
		}

		instance.Fields["_wheres"] = wheres
		instance.Fields["_bindings"] = bindings
		return instance

	case "get":
		if r.DB == nil {
			fmt.Println("[GranMySQL] Error: No DB connection")
			return "[]"
		}

		table := instance.Fields["_table"].(string)
		sel := instance.Fields["_select"].(string)
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		query := fmt.Sprintf("SELECT %s FROM %s", sel, table)
		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}

		// Reset state after query build
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_select"] = "*"
		// Keep table? Usually builder resets, but let's keep it simple.

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			fmt.Printf("[GranMySQL] Error en get: %v\n", err)
			return "[]"
		}
		defer rows.Close()

		return rowsToJSON(rows)

	case "first":
		// Similar to get but returns single object or null
		// Add LIMIT 1
		if r.DB == nil {
			return nil
		}

		table := instance.Fields["_table"].(string)
		sel := instance.Fields["_select"].(string)
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		query := fmt.Sprintf("SELECT %s FROM %s", sel, table)
		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}
		query += " LIMIT 1"

		// Reset state
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			return nil
		}
		defer rows.Close()

		jsonStr := rowsToJSON(rows)
		// If "[]", return nil/false? Or the first element?
		// rowsToJSON returns string "[]" or "[{...}]"
		if jsonStr == "[]" {
			return nil
		}
		// Extract first object from JSON array string (hacky but consistent with current string returns)
		// Better: return the raw map if possible? Joss uses strings mostly for now.
		// Let's return the string of the object.
		return strings.TrimSuffix(strings.TrimPrefix(jsonStr.(string), "["), "]")

	case "insert":
		if r.DB == nil {
			return false
		}
		if len(args) > 0 {
			// Support insert(["col1", "col2"], ["val1", "val2"])
			if len(args) == 2 {
				cols := args[0].([]interface{})
				vals := args[1].([]interface{})

				if len(cols) != len(vals) {
					return false
				}

				table := instance.Fields["_table"].(string)
				colNames := []string{}
				placeholders := []string{}
				bindings := []interface{}{}

				for _, c := range cols {
					colNames = append(colNames, fmt.Sprintf("%v", c))
					placeholders = append(placeholders, "?")
				}
				for _, v := range vals {
					bindings = append(bindings, v)
				}

				query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))

				_, err := r.DB.Exec(query, bindings...)
				if err != nil {
					fmt.Printf("[GranMySQL] Error insert: %v\n", err)
					return false
				}
				return true
			}
		}
		return false

	case "query":
		// Raw query support (still needed for migrations)
		if len(args) > 0 {
			if sqlStr, ok := args[0].(string); ok {
				if r.DB == nil {
					return nil
				}
				_, err := r.DB.Exec(sqlStr)
				if err != nil {
					fmt.Printf("[GranMySQL] Error query: %v\n", err)
					return false
				}
				return true
			}
		}

	// Legacy support (tabla, comparar, comparable) -> Map to builder
	case "legacy_where":
		// ...
	}
	return nil
}

func rowsToJSON(rows *sql.Rows) interface{} {
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
			rowStr += fmt.Sprintf("\"%s\": \"%v\"", colName, valVal)
			if i < len(cols)-1 {
				rowStr += ", "
			}
		}
		rowStr += "}"
		results = append(results, rowStr)
	}
	return "[" + strings.Join(results, ", ") + "]"
}

// Auth Implementation
func (r *Runtime) executeAuthMethod(instance *Instance, method string, args []interface{}) interface{} {
	prefix := getTablePrefix()
	usersTable := prefix + "users"

	// Auto-migrate check (lazy)
	if _, ok := instance.Fields["_migrated"]; !ok {
		if r.DB != nil {
			// Check if table exists
			// Simplified: Just try create if not exists
			query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255),
				email VARCHAR(255) UNIQUE,
				password VARCHAR(255),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			)`, usersTable)
			r.DB.Exec(query)
			instance.Fields["_migrated"] = true
		}
	}

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

				hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return false
				}
				hashedPassword := string(hashedBytes)

				if r.DB == nil {
					return false
				}

				query := fmt.Sprintf("INSERT INTO %s (name, email, password, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", usersTable)
				_, err = r.DB.Exec(query, name, email, hashedPassword)
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
			var userId int
			var userName string
			query := fmt.Sprintf("SELECT id, name, password FROM %s WHERE email = ?", usersTable)
			err := r.DB.QueryRow(query, email).Scan(&userId, &userName, &storedHash)
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

			// Generate JWT token with 3 months expiration (initial login)
			return r.generateJWT(userId, email, userName, false)

		}
	case "refresh":
		// Refresh JWT token with 6 months expiration
		if len(args) >= 1 {
			tokenString := args[0].(string)

			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "joss_default_secret_change_in_production"
			}

			// Parse and validate existing token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				fmt.Printf("[Security] Token inválido: %v\n", err)
				return false
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				fmt.Println("[Security] Error parseando claims")
				return false
			}

			// Extract user info from claims
			userIdFloat, _ := claims["user_id"].(float64)
			userId := int(userIdFloat)
			email, _ := claims["email"].(string)
			userName, _ := claims["name"].(string)

			fmt.Println("[Security] Token refrescado exitosamente.")
			// Generate new token with 6 months expiration (refresh)
			return r.generateJWT(userId, email, userName, true)
		}
	}
	return nil
}

// generateJWT creates a JWT token with configurable expiration
func (r *Runtime) generateJWT(userId int, email string, userName string, isRefresh bool) interface{} {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	// Get expiration time from env or use defaults
	var expirationMonths int
	if isRefresh {
		// Refresh token: 6 months default
		refreshMonths := os.Getenv("JWT_REFRESH_EXPIRY_MONTHS")
		if refreshMonths != "" {
			fmt.Sscanf(refreshMonths, "%d", &expirationMonths)
		} else {
			expirationMonths = 6
		}
	} else {
		// Initial token: 3 months default
		initialMonths := os.Getenv("JWT_INITIAL_EXPIRY_MONTHS")
		if initialMonths != "" {
			fmt.Sscanf(initialMonths, "%d", &expirationMonths)
		} else {
			expirationMonths = 3
		}
	}

	// Calculate expiration time (approximate: 30 days per month)
	expirationDuration := time.Duration(expirationMonths) * 30 * 24 * time.Hour

	// Create claims
	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"name":    userName,
		"exp":     time.Now().Add(expirationDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("[Security] Error generando JWT: %v\n", err)
		return false
	}

	return tokenString
}
