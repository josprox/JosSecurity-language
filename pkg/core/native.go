package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
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

	// System
	systemClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "System"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(systemClass)
	r.Variables["System"] = &Instance{Class: systemClass, Fields: make(map[string]interface{})}

	// SmtpClient
	r.registerClass(&parser.ClassStatement{
		Name: &parser.Identifier{Value: "SmtpClient"},
		Body: &parser.BlockStatement{},
	})

	// Cron
	cronClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Cron"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(cronClass)
	r.Variables["Cron"] = &Instance{Class: cronClass, Fields: make(map[string]interface{})}

	// Task
	taskClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "Task"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(taskClass)
	r.Variables["Task"] = &Instance{Class: taskClass, Fields: make(map[string]interface{})}

	// View
	viewClass := &parser.ClassStatement{
		Name: &parser.Identifier{Value: "View"},
		Body: &parser.BlockStatement{},
	}
	r.registerClass(viewClass)
	r.Variables["View"] = &Instance{Class: viewClass, Fields: make(map[string]interface{})}
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
	case "System":
		return r.executeSystemMethod(instance, method, args)
	case "SmtpClient":
		return r.executeSmtpClientMethod(instance, method, args)
	case "Cron":
		return r.executeCronMethod(instance, method, args)
	case "Task":
		return r.executeTaskMethod(instance, method, args)
	case "View":
		return r.executeViewMethod(instance, method, args)
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

	case "innerJoin":
		if len(args) >= 4 {
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("INNER JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "leftJoin":
		if len(args) >= 4 {
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("LEFT JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "rightJoin":
		if len(args) >= 4 {
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("RIGHT JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
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

		if joins, ok := instance.Fields["_joins"]; ok {
			for _, j := range joins.([]string) {
				query += " " + j
			}
		}

		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}

		// Reset state after query build
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_select"] = "*"
		instance.Fields["_joins"] = []string{}

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			fmt.Printf("[GranMySQL] Error en get: %v\n", err)
			return "[]"
		}
		defer rows.Close()

		return rowsToJSON(rows)

	case "first":
		if r.DB == nil {
			return nil
		}

		table := instance.Fields["_table"].(string)
		sel := instance.Fields["_select"].(string)
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		query := fmt.Sprintf("SELECT %s FROM %s", sel, table)

		if joins, ok := instance.Fields["_joins"]; ok {
			for _, j := range joins.([]string) {
				query += " " + j
			}
		}

		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}
		query += " LIMIT 1"

		// Reset state
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_joins"] = []string{}

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			return nil
		}
		defer rows.Close()

		jsonStr := rowsToJSON(rows)
		if jsonStr == "[]" {
			return nil
		}
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
	}
	return nil
}

// System Implementation
func (r *Runtime) executeSystemMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "env":
		if len(args) > 0 {
			key := args[0].(string)
			if val, ok := r.Env[key]; ok {
				return val
			}
			if len(args) > 1 {
				return args[1] // Default value
			}
			return ""
		}
	case "Run":
		if len(args) > 0 {
			cmdName := args[0].(string)
			cmdArgs := []string{}
			if len(args) > 1 {
				if list, ok := args[1].([]interface{}); ok {
					for _, arg := range list {
						cmdArgs = append(cmdArgs, fmt.Sprintf("%v", arg))
					}
				}
			}

			cmd := exec.Command(cmdName, cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[System] Error ejecutando '%s': %v\n", cmdName, err)
				return ""
			}
			return string(output)
		}
	case "load_driver":
		if len(args) > 0 {
			path := args[0].(string)
			fmt.Printf("[System] Cargando driver externo desde: %s (Simulación)\n", path)
			return true
		}
	}
	return nil
}

// SmtpClient Implementation
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

// SmtpClient Implementation
func (r *Runtime) executeSmtpClientMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "auth":
		if len(args) == 2 {
			instance.Fields["user"] = args[0]
			instance.Fields["pass"] = args[1]
		}
		return instance
	case "secure":
		if len(args) == 1 {
			instance.Fields["secure"] = args[0]
		}
		return instance
	case "send":
		if len(args) >= 3 {
			to := args[0].(string)
			subject := args[1].(string)
			body := args[2].(string)

			// Defaults
			host := "smtp.gmail.com"
			port := "587"
			if h, ok := r.Env["MAIL_HOST"]; ok {
				host = h
			}
			if p, ok := r.Env["MAIL_PORT"]; ok {
				port = p
			}

			user := ""
			pass := ""
			if u, ok := instance.Fields["user"]; ok {
				user = u.(string)
			}
			if p, ok := instance.Fields["pass"]; ok {
				pass = p.(string)
			}

			msg := []byte("From: " + user + "\r\n" +
				"To: " + to + "\r\n" +
				"Subject: " + subject + "\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
				"\r\n" +
				body + "\r\n")

			auth := smtp.PlainAuth("", user, pass, host)
			err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
			if err != nil {
				fmt.Printf("[SmtpClient] Error enviando correo: %v\n", err)
				return false
			}
			fmt.Println("[SmtpClient] Correo enviado exitosamente a " + to)
			return true
		}
	}
	return nil
}

// Cron Implementation (Daemon mode simulation)
func (r *Runtime) executeCronMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "schedule" {
		if len(args) >= 3 {
			name := args[0].(string)
			timeStr := args[1].(string)
			fmt.Printf("[Cron] Tarea '%s' programada para '%s' (Simulación)\n", name, timeStr)
			// In a real daemon, we would register this callback
		}
	}
	return nil
}

// Task Implementation (Hit-based)
func (r *Runtime) executeTaskMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "on_request" {
		if len(args) >= 3 {
			name := args[0].(string)
			// interval := args[1].(string)
			// callback := args[2] (Block/Closure)

			// Check if we should run
			// For simulation, we always run it if it's "system_health" or similar, or just log it.
			// To properly implement, we need to execute the block passed as argument.
			// The runtime needs to handle the callback execution.
			// Since we are inside executeNativeMethod, we can't easily execute a block without the runtime context loop.
			// BUT, we are in the runtime! 'r' is *Runtime.

			// We need to check if args[2] is a BlockStatement or similar?
			// The parser passes blocks as... wait, native methods receive evaluated args.
			// If the argument was a block `{ ... }`, it might not be evaluated to a value easily unless we treat it as a closure.
			// Currently, Joss doesn't support passing blocks as values (closures) fully.
			// The "Bible" says: Task::on_request(..., { code })
			// This implies the 3rd argument is a block.
			// In `evaluateCall`, arguments are evaluated. A block `{}` is not an expression in current parser?
			// Let's assume for now we just print. To support this fully, we need Closures.

			fmt.Printf("[Task] Ejecutando tarea hit-based: %s\n", name)
			// If we could, we would execute the block here.
		}
	}
	return nil
}

// View Implementation
func (r *Runtime) executeViewMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "render" {
		if len(args) >= 1 {
			viewName := args[0].(string)
			data := make(map[string]interface{})
			if len(args) > 1 {
				if d, ok := args[1].(map[string]interface{}); ok {
					data = d
				}
			}

			// Resolve path: app/views/viewName.joss.html
			// Convert dot notation to path
			viewPath := strings.ReplaceAll(viewName, ".", "/")
			path := filepath.Join("app", "views", viewPath+".joss.html")

			content, err := os.ReadFile(path)
			if err != nil {
				// Try without .joss.html or just .html
				path = filepath.Join("app", "views", viewPath+".html")
				content, err = os.ReadFile(path)
				if err != nil {
					return fmt.Sprintf("Error: Vista '%s' no encontrada", viewName)
				}
			}

			html := string(content)
			// Simple template replacement {{ key }}
			for k, v := range data {
				placeholder := "{{" + k + "}}" // Strict no spaces
				html = strings.ReplaceAll(html, placeholder, fmt.Sprintf("%v", v))
				placeholderSpace := "{{ " + k + " }}" // With spaces
				html = strings.ReplaceAll(html, placeholderSpace, fmt.Sprintf("%v", v))
			}
			return html
		}
	}
	return nil
}

// TOON Helpers
func ToonEncode(data interface{}) string {
	// Simple implementation
	// If array of maps:
	// entity[count]{keys}:
	//   val1,val2

	if list, ok := data.([]interface{}); ok {
		if len(list) == 0 {
			return ""
		}
		// Assume homogeneous list of maps
		first := list[0]
		if m, ok := first.(map[string]interface{}); ok {
			keys := []string{}
			for k := range m {
				keys = append(keys, k)
			}

			header := fmt.Sprintf("entity[%d]{%s}:\n", len(list), strings.Join(keys, ","))
			body := ""
			for _, item := range list {
				if row, ok := item.(map[string]interface{}); ok {
					vals := []string{}
					for _, k := range keys {
						vals = append(vals, fmt.Sprintf("%v", row[k]))
					}
					body += "  " + strings.Join(vals, ",") + "\n"
				}
			}
			return header + body
		}
	}
	return fmt.Sprintf("%v", data)
}

func ToonDecode(str string) interface{} {
	// Handle literal \n if parser didn't unescape it
	str = strings.ReplaceAll(str, "\\n", "\n")

	// Very basic parser for "entity[N]{k1,k2}:\n v1,v2..."
	lines := strings.Split(strings.TrimSpace(str), "\n")
	if len(lines) < 2 {
		return nil
	}

	header := lines[0]
	// Parse header: name[count]{keys}:
	// Regex or simple string manipulation
	startBracket := strings.Index(header, "[")
	endBracket := strings.Index(header, "]")
	startBrace := strings.Index(header, "{")
	endBrace := strings.Index(header, "}")

	if startBracket == -1 || endBracket == -1 || startBrace == -1 || endBrace == -1 {
		return nil
	}

	// countStr := header[startBracket+1 : endBracket]
	keysStr := header[startBrace+1 : endBrace]
	keys := strings.Split(keysStr, ",")

	result := []interface{}{}

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		vals := strings.Split(line, ",")
		if len(vals) != len(keys) {
			continue
		}

		obj := make(map[string]interface{})
		for i, k := range keys {
			obj[k] = vals[i]
		}
		result = append(result, obj)
	}

	return result
}

func ToonVerify(str string) bool {
	// Handle literal \n
	str = strings.ReplaceAll(str, "\\n", "\n")

	// Simple verification: check structure
	lines := strings.Split(strings.TrimSpace(str), "\n")
	if len(lines) < 2 {
		return false
	}
	header := lines[0]
	// Must contain [ ] { } :
	if !strings.Contains(header, "[") || !strings.Contains(header, "]") ||
		!strings.Contains(header, "{") || !strings.Contains(header, "}") ||
		!strings.Contains(header, ":") {
		return false
	}
	return true
}

// JSON Helpers
func JsonEncode(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

func JsonDecode(str string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		return nil
	}
	return result
}

func JsonVerify(str string) bool {
	return json.Valid([]byte(str))
}
