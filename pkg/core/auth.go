package core

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"strings" // Added for TrimSpace

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Auth Implementation
func (r *Runtime) executeAuthMethod(instance *Instance, method string, args []interface{}) interface{} {
	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	usersTable := prefix + "users"
	rolesTable := prefix + "roles"

	fmt.Printf("[Auth Debug] Prefix: '%s', Users Table: '%s'\n", prefix, usersTable)

	// Asegurar que las tablas y columnas existan (Auto-Migración)
	r.ensureAuthTables(usersTable, rolesTable)

	switch method {
	case "create":
		// Auth::create({ ... })
		if len(args) > 0 {
			if data, ok := args[0].(map[string]interface{}); ok {
				// 1. Generar User Token Obligatorio
				userToken := uuid.New().String()

				// Definir función de tiempo según DB
				nowFunc := "NOW()"
				if val, ok := r.Env["DB"]; ok && val == "sqlite" {
					nowFunc = "CURRENT_TIMESTAMP"
				}

				// Extraer campos (Sin 'name')
				username := getString(data, "username", "")
				firstName := getString(data, "first_name", "")
				lastName := getString(data, "last_name", "")
				email := strings.TrimSpace(getString(data, "email", "")) // Trim Email
				phone := getString(data, "phone", "")
				password := getString(data, "password", "")

				// Opcional: role_id
				roleId := 2
				if rId, ok := data["role_id"].(int64); ok {
					roleId = int(rId)
				}

				hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return false
				}
				hashedPassword := string(hashedBytes)

				if r.DB == nil {
					return false
				}

				// Query explícito
				query := fmt.Sprintf(`INSERT INTO %s 
					(user_token, username, first_name, last_name, email, phone, password, role_id, created_at, updated_at, verificado) 
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, %s, %s, 0)`, usersTable, nowFunc, nowFunc)

				_, err = r.DB.Exec(query, userToken, username, firstName, lastName, email, phone, hashedPassword, roleId)
				if err != nil {
					fmt.Printf("[Security] Error creando usuario: %v\n", err)
					return false
				}
				fmt.Println("[Security] Usuario registrado exitosamente.")
				return userToken
			}
		}

	case "attempt":
		if len(args) >= 2 {
			if args[0] == nil || args[1] == nil {
				LogError("[Auth] Attempt failed: Email or Password is nil")
				return false
			}
			email := strings.TrimSpace(args[0].(string)) // Trim Email
			password := args[1].(string)

			if r.DB == nil {
				LogError("[Auth] Attempt failed: Database connection is nil")
				return false
			}

			// Variables para Scan
			var storedHash sql.NullString
			var userId int
			var userName sql.NullString // Username del sistema
			var userToken sql.NullString
			var roleName sql.NullString
			var verificado int

			// Join con roles
			query := fmt.Sprintf(`
				SELECT u.id, u.user_token, u.username, u.password, u.verificado, r.name 
				FROM %s u 
				LEFT JOIN %s r ON u.role_id = r.id 
				WHERE u.email = ?`, usersTable, rolesTable)

			err := r.DB.QueryRow(query, email).Scan(&userId, &userToken, &userName, &storedHash, &verificado, &roleName)
			if err != nil {
				if err == sql.ErrNoRows {
					LogError("[Auth] User not found for email: '%s'", email)
				} else {
					LogError("[Auth] Database error looking up '%s': %v", email, err)
				}
				return false
			}

			if verificado == 0 {
				LogError("[Auth] Account not verified for '%s'", email)
				return false // Fallar si no está verificado
			}

			err = bcrypt.CompareHashAndPassword([]byte(storedHash.String), []byte(password))
			if err != nil {
				LogError("[Auth] Password mismatch for '%s'", email)
				return false
			}

			LogInfo("[Auth] Login successful for '%s' (ID: %d)", email, userId)

			// Guardar en Sesión ($__session)
			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					sessInst.Fields["user_id"] = userId
					sessInst.Fields["user_token"] = userToken.String
					sessInst.Fields["user_name"] = userName.String
					sessInst.Fields["user_email"] = email
					sessInst.Fields["user_role"] = roleName.String
					sessInst.Fields["last_login_at"] = time.Now().Format("2006-01-02 15:04:05")
				}
			}

			// Actualizar last_login_at
			updateQuery := fmt.Sprintf("UPDATE %s SET last_login_at = %s WHERE id = ?", usersTable, "CURRENT_TIMESTAMP")
			if val, ok := r.Env["DB"]; ok && val == "mysql" {
				updateQuery = fmt.Sprintf("UPDATE %s SET last_login_at = NOW() WHERE id = ?", usersTable)
			}
			r.DB.Exec(updateQuery, userId)

			// Retornar JWT Token
			return r.generateJWT(userId, email, userName.String, roleName.String, false)
		}

	case "check":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if _, ok := sessInst.Fields["user_id"]; ok {
					return true
				}
			}
		}
		return false

	case "verify":
		if len(args) == 1 {
			token := args[0].(string)
			if r.DB == nil {
				return false
			}
			var id int
			query := fmt.Sprintf("SELECT id FROM %s WHERE user_token = ? LIMIT 1", usersTable)
			err := r.DB.QueryRow(query, token).Scan(&id)
			if err != nil {
				return false
			}
			update := fmt.Sprintf("UPDATE %s SET verificado = 1 WHERE id = ?", usersTable)
			_, err = r.DB.Exec(update, id)
			return err == nil
		}

	case "user":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if uid, ok := sessInst.Fields["user_id"]; ok {
					if r.DB == nil {
						return nil
					}

					// Objeto usuario a retornar
					user := make(map[string]interface{})

					var id, roleId int
					var username, email, firstName, lastName, userToken, createdAt sql.NullString
					var pPhone sql.NullString

					query := fmt.Sprintf(`SELECT id, username, first_name, last_name, email, phone, role_id, user_token, created_at FROM %s WHERE id = ?`, usersTable)

					err := r.DB.QueryRow(query, uid).Scan(&id, &username, &firstName, &lastName, &email, &pPhone, &roleId, &userToken, &createdAt)
					if err != nil {
						fmt.Printf("[Auth Error] User Query Failed for ID %v: %v\n", uid, err)
					}
					if err == nil {
						user["id"] = id
						user["username"] = username.String
						user["first_name"] = firstName.String
						user["last_name"] = lastName.String
						// Helper name para UI (concatenado)
						user["full_name"] = firstName.String + " " + lastName.String
						user["email"] = email.String
						user["phone"] = pPhone.String
						user["role_id"] = roleId
						user["user_token"] = userToken.String
						user["created_at"] = createdAt.String
						// Compatibility for templates using user.name
						user["name"] = firstName.String

						// Debug Print
						fmt.Printf("[Auth] User found: %s (ID: %d)\n", firstName.String, id)

						return &Instance{
							Fields: user,
						}
					}
				}
			}
		}
		return nil

	case "guest":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if _, ok := sessInst.Fields["user_id"]; ok {
					return false
				}
			}
		}
		return true

	case "hasRole":
		if len(args) == 1 {
			roleToCheck := args[0].(string)
			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					if currentRole, ok := sessInst.Fields["user_role"]; ok {
						if currentRole == roleToCheck {
							return true
						}
						// Admin bypass
						if currentRole == "admin" {
							return true
						}
					}
				}
			}
		}
		return false

	case "refresh":
		if len(args) == 1 {
			if id, ok := args[0].(int); ok {
				var email, username, roleName string
				// Need to join with roles to get role name
				prefix := "js_"
				if val, ok := r.Env["PREFIX"]; ok {
					prefix = val
				}
				usersTable := prefix + "users"
				rolesTable := prefix + "roles"

				// Fixed query to include role
				query := fmt.Sprintf(`
					SELECT u.email, u.username, r.name 
					FROM %s u 
					LEFT JOIN %s r ON u.role_id = r.id 
					WHERE u.id = ?`, usersTable, rolesTable)

				err := r.DB.QueryRow(query, id).Scan(&email, &username, &roleName)
				if err != nil {
					return false
				}
				return r.generateJWT(id, email, username, roleName, false)
			}
		}

	case "update":
		if len(args) == 2 {
			id, ok1 := args[0].(int)
			data, ok2 := args[1].(map[string]interface{})

			if ok1 && ok2 {
				if r.DB == nil {
					return false
				}

				// Handle Password Hashing
				if pwd, ok := data["password"]; ok {
					passwordStr := fmt.Sprintf("%v", pwd)
					if passwordStr != "" {
						hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordStr), bcrypt.DefaultCost)
						if err == nil {
							data["password"] = string(hashedBytes)
						}
					} else {
						// Don't update empty password
						delete(data, "password")
					}
				}

				// Construct Query
				var sets []string
				var vals []interface{}

				// Add updated_at automatically
				if val, ok := r.Env["DB"]; ok && val == "sqlite" {
					sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
				} else {
					sets = append(sets, "updated_at = NOW()")
				}

				for k, v := range data {
					// Protect strict columns if needed, but for now trust controller
					if k != "id" && k != "user_token" && k != "created_at" && k != "updated_at" {
						sets = append(sets, fmt.Sprintf("%s = ?", k))
						vals = append(vals, v)
					}
				}
				vals = append(vals, id)

				if len(sets) == 0 {
					return true // Nothing to update
				}

				query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", usersTable, strings.Join(sets, ", "))
				_, err := r.DB.Exec(query, vals...)
				return err == nil
			}
		}

	case "delete":
		if len(args) == 1 {
			if id, ok := args[0].(int); ok {
				if r.DB == nil {
					return false
				}
				query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", usersTable)
				_, err := r.DB.Exec(query, id)
				return err == nil
			}
		}

	case "logout":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				delete(sessInst.Fields, "user_id")
				delete(sessInst.Fields, "user_token")
				delete(sessInst.Fields, "user_name")
				delete(sessInst.Fields, "user_email")
				delete(sessInst.Fields, "user_role")
				delete(sessInst.Fields, "last_login_at")
			}
		}
		return true

	case "validateToken":
		if len(args) == 1 {
			tokenString := args[0].(string)
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

			claims, valid := r.ValidateJWT(tokenString)
			if valid {
				if sessVal, ok := r.Variables["$__session"]; ok {
					if sessInst, ok := sessVal.(*Instance); ok {
						sessInst.Fields["user_id"] = int(claims["user_id"].(float64))
						sessInst.Fields["user_email"] = claims["email"]
						sessInst.Fields["user_name"] = claims["name"]
					}
				}
				return true
			}
			return false
		}
	}
	return nil
}

// --- HELPERS Y CONFIGURACIÓN DE TABLAS ---

var authTablesEnsured bool

func (r *Runtime) ensureAuthTables(usersTable, rolesTable string) {
	if r.DB == nil || authTablesEnsured {
		return
	}

	// 1. Crear Tabla Roles
	createRoles := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(50) NOT NULL UNIQUE
	);`, rolesTable)

	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		createRoles = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE
		);`, rolesTable)
	}
	r.DB.Exec(createRoles)

	// 2. Crear Tabla Users (Sin columna 'name')
	createUsers := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_token VARCHAR(128) NOT NULL,
		username VARCHAR(50) NOT NULL,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		phone VARCHAR(20),
		password VARCHAR(255) NOT NULL,
		role_id INTEGER NOT NULL DEFAULT 2,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		verificado INTEGER DEFAULT 0,
		last_login_at DATETIME,
		FOREIGN KEY(role_id) REFERENCES %s(id)
	);`, usersTable, rolesTable)

	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		createUsers = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_token VARCHAR(128) NOT NULL,
			username VARCHAR(50) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL UNIQUE,
			phone VARCHAR(20),
			password VARCHAR(255) NOT NULL,
			role_id INT NOT NULL DEFAULT 2,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			verificado TINYINT(1) DEFAULT 0,
			last_login_at DATETIME,
			FOREIGN KEY(role_id) REFERENCES %s(id)
		);`, usersTable, rolesTable)
	}
	r.DB.Exec(createUsers)

	// 3. Insertar Roles por defecto
	r.DB.Exec(fmt.Sprintf("INSERT OR IGNORE INTO %s (name) VALUES ('admin'), ('client')", rolesTable))
	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		r.DB.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ('admin'), ('client') ON DUPLICATE KEY UPDATE name=name", rolesTable))
	}

	authTablesEnsured = true

	// 4. AUTO-MIGRACIÓN (Esto arregla el problema de SQLite)
	isMySQL := false
	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		isMySQL = true
	}

	// Agregamos columnas si no existen (Patching)
	patchColumn(r.DB, usersTable, "username", "VARCHAR(50) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "user_token", "VARCHAR(128) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "first_name", "VARCHAR(100) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "last_name", "VARCHAR(100) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "phone", "VARCHAR(20) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "verificado", "INTEGER DEFAULT 0", isMySQL)
	patchColumn(r.DB, usersTable, "created_at", "DATETIME DEFAULT CURRENT_TIMESTAMP", isMySQL)
	patchColumn(r.DB, usersTable, "updated_at", "DATETIME DEFAULT CURRENT_TIMESTAMP", isMySQL)
	patchColumn(r.DB, usersTable, "last_login_at", "DATETIME", isMySQL)
}

func patchColumn(db *sql.DB, table, col, def string, isMySQL bool) {
	// Verificar si la columna ya existe
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s LIMIT 1", col, table))
	if err == nil {
		rows.Close()
		return // Existe, no hacemos nada
	}

	// Si falla, asumimos que no existe y la creamos
	fmt.Printf("[Auth] Auto-patching: Agregando columna '%s' a tabla '%s'...\n", col, table)

	alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, def)
	if isMySQL {
		alter = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, def)
	}
	_, err = db.Exec(alter)
	// Ignoramos error si falla el alter, para no detener el runtime
}

func getString(data map[string]interface{}, key, def string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return def
}

func (r *Runtime) generateJWT(userId int, email string, userName string, roleName string, isRefresh bool) interface{} {
	expirationTime := time.Now().Add(24 * 30 * time.Hour)
	if isRefresh {
		expirationTime = time.Now().Add(24 * 180 * time.Hour)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"name":    userName,
		"role":    roleName,
		"exp":     expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		fmt.Printf("[Security] Error generando JWT: %v\n", err)
		return false
	}

	return tokenString
}

func (r *Runtime) ValidateJWT(tokenString string) (map[string]interface{}, bool) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "joss_default_secret_change_in_production"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	}

	return nil, false
}
