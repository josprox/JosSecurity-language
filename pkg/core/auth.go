package core

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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

	// Ensure tables exist
	r.ensureAuthTables(usersTable, rolesTable)

	switch method {
	case "create":
		// Auth::create({ ... })
		if len(args) > 0 {
			if data, ok := args[0].(map[string]interface{}); ok {
				// Generate Defaults
				userToken := uuid.New().String()
				nowFunc := "NOW()"
				if val, ok := r.Env["DB"]; ok && val == "sqlite" {
					nowFunc = "CURRENT_TIMESTAMP"
				}

				// Extract fields with defaults
				username := getString(data, "username", "")
				firstName := getString(data, "first_name", "")
				lastName := getString(data, "last_name", "")
				email := getString(data, "email", "")
				phone := getString(data, "phone", "")
				password := getString(data, "password", "")

				// Optional: role_id
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

				query := fmt.Sprintf(`INSERT INTO %s 
					(user_token, username, first_name, last_name, email, phone, password, role_id, created_at, updated_at, verificado) 
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, %s, %s, 0)`, usersTable, nowFunc, nowFunc)

				_, err = r.DB.Exec(query, userToken, username, firstName, lastName, email, phone, hashedPassword, roleId)
				if err != nil {
					fmt.Printf("[Security] Error creando usuario: %v\n", err)
					return false
				}
				fmt.Println("[Security] Usuario registrado exitosamente.")
				fmt.Println("[Security] Usuario registrado exitosamente.")
				return userToken
			}
		}
	case "attempt":
		if len(args) >= 2 {
			if args[0] == nil || args[1] == nil {
				return false
			}
			email := args[0].(string)
			password := args[1].(string)

			if r.DB == nil {
				return false
			}

			var storedHash string
			var userId int
			var userName string
			var userToken string
			var roleName sql.NullString
			var verificado int

			// Join with roles table to get role name
			query := fmt.Sprintf(`
				SELECT u.id, u.user_token, u.username, u.password, u.verificado, r.name 
				FROM %s u 
				LEFT JOIN %s r ON u.role_id = r.id 
				WHERE u.email = ?`, usersTable, rolesTable)

			err := r.DB.QueryRow(query, email).Scan(&userId, &userToken, &userName, &storedHash, &verificado, &roleName)
			if err != nil {
				if err == sql.ErrNoRows {
					fmt.Println("[Security] Usuario no encontrado.")
				} else {
					fmt.Printf("[Security] Error DB: %v\n", err)
				}
				return false
			}

			if verificado == 0 {
				fmt.Println("[Security] Cuenta no verificada.")
				return false // Should return false to fail the check in templates
			}

			err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
			if err != nil {
				fmt.Println("[Security] Contraseña incorrecta.")
				return false
			}

			fmt.Println("[Security] Login exitoso.")

			// Store in session (for stateful support)
			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					sessInst.Fields["user_id"] = userId
					sessInst.Fields["user_token"] = userToken
					sessInst.Fields["user_name"] = userName
					sessInst.Fields["user_email"] = email
					sessInst.Fields["user_role"] = roleName.String
					sessInst.Fields["last_login_at"] = time.Now().Format("2006-01-02 15:04:05")
				}
			}

			// Update last_login_at
			updateQuery := fmt.Sprintf("UPDATE %s SET last_login_at = %s WHERE id = ?", usersTable, "CURRENT_TIMESTAMP") // simplified for sqlite/mysql compat
			if val, ok := r.Env["DB"]; ok && val == "mysql" {
				updateQuery = fmt.Sprintf("UPDATE %s SET last_login_at = NOW() WHERE id = ?", usersTable)
			}
			r.DB.Exec(updateQuery, userId)

			// Return JWT Token (Compliance with Bible)
			return r.generateJWT(userId, email, userName, false)
		}
	// ... other cases
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

			// Check if token exists
			var id int
			query := fmt.Sprintf("SELECT id FROM %s WHERE user_token = ? LIMIT 1", usersTable)
			err := r.DB.QueryRow(query, token).Scan(&id)
			if err != nil {
				return false
			}

			// Update verified status
			update := fmt.Sprintf("UPDATE %s SET verificado = 1 WHERE id = ?", usersTable)
			_, err = r.DB.Exec(update, id)
			return err == nil
		}

	case "user":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if uid, ok := sessInst.Fields["user_id"]; ok {
					// Fetch full user from DB
					if r.DB == nil {
						return nil
					}

					// Return map
					user := make(map[string]interface{})
					var id, roleId int
					var username, email, firstName, lastName, userToken string
					var pPhone sql.NullString

					query := fmt.Sprintf(`SELECT id, username, first_name, last_name, email, phone, role_id, user_token FROM %s WHERE id = ?`, usersTable)

					err := r.DB.QueryRow(query, uid).Scan(&id, &username, &firstName, &lastName, &email, &pPhone, &roleId, &userToken)
					if err == nil {
						user["id"] = id
						user["username"] = username
						user["first_name"] = firstName
						user["last_name"] = lastName
						// Helper name for UI
						user["name"] = firstName + " " + lastName
						user["email"] = email
						user["phone"] = pPhone.String
						user["role_id"] = roleId
						user["user_token"] = userToken

						return user
					}
				}
			}
		}
		return nil

	case "guest":
		// Inverse of check
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

	case "logout":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				// Clear session fields
				delete(sessInst.Fields, "user_id")
				delete(sessInst.Fields, "user_token")
				delete(sessInst.Fields, "user_name")
				delete(sessInst.Fields, "user_email")
				delete(sessInst.Fields, "user_role")
				delete(sessInst.Fields, "last_login_at")
			}
		}
		return true
	}
	return nil
}

// Avoid repeated checks
var authTablesEnsured bool

func (r *Runtime) ensureAuthTables(usersTable, rolesTable string) {
	if r.DB == nil || authTablesEnsured {
		return
	}

	// Roles Table
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

	// Users Table (17 Fields)
	// We use "password" instead of "contra" for standardization.
	createUsers := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_token VARCHAR(128) NOT NULL,
		username VARCHAR(50) NOT NULL,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		email VARCHAR(150) NOT NULL UNIQUE,
		phone VARCHAR(20),
		last_ip VARCHAR(45),
		last_user_agent VARCHAR(255),
		last_login_at TIMESTAMP NULL,
		last_refresh_at TIMESTAMP NULL,
		last_logout_at TIMESTAMP NULL,
		last_seen_at TIMESTAMP NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
		verificado TINYINT(1) DEFAULT 0,
		role_id INTEGER DEFAULT 2
	);`, usersTable)

	isMySQL := false
	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		isMySQL = true
		createUsers = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id BIGINT(20) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_token VARCHAR(128) NOT NULL,
			username VARCHAR(50) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			email VARCHAR(150) NOT NULL,
			phone VARCHAR(20) NULL,
			last_ip VARCHAR(45) NULL,
			last_user_agent VARCHAR(255) NULL,
			last_login_at TIMESTAMP NULL,
			last_refresh_at TIMESTAMP NULL,
			last_logout_at TIMESTAMP NULL,
			last_seen_at TIMESTAMP NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NULL DEFAULT NULL,
			updated_at TIMESTAMP NULL DEFAULT NULL,
			verificado TINYINT(1) NOT NULL DEFAULT 0,
			role_id INT DEFAULT 2
		);`, usersTable)
	}
	r.DB.Exec(createUsers)

	// Self-Healing: Check for missing columns (e.g. if table existed with old schema or alternate naming)
	patchColumn(r.DB, usersTable, "user_token", "VARCHAR(128) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "username", "VARCHAR(50) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "first_name", "VARCHAR(100) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "last_name", "VARCHAR(100) NOT NULL DEFAULT ''", isMySQL)
	patchColumn(r.DB, usersTable, "phone", "VARCHAR(20)", isMySQL)
	patchColumn(r.DB, usersTable, "password", "VARCHAR(255) NOT NULL DEFAULT ''", isMySQL) // Ensure password exists
	patchColumn(r.DB, usersTable, "verificado", "TINYINT(1) DEFAULT 0", isMySQL)
	patchColumn(r.DB, usersTable, "role_id", "INTEGER DEFAULT 2", isMySQL)
	patchColumn(r.DB, usersTable, "created_at", "TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP", isMySQL)
	patchColumn(r.DB, usersTable, "updated_at", "TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP", isMySQL)

	// Seed Roles
	var count int
	r.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", rolesTable)).Scan(&count)
	if count == 0 {
		r.DB.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES ('admin'), ('client')", rolesTable))
	}

	authTablesEnsured = true
}

// patchColumn adds a column if it doesn't exist
func patchColumn(db *sql.DB, table, col, def string, isMySQL bool) {
	// Simple check: SELECT col FROM table LIMIT 1
	// Usage: SELECT column_name FROM table_name
	// But to check existence reliably across DBs without SELECT * overhead:
	// We simply try to select the specific column.
	rows, err := db.Query(fmt.Sprintf("SELECT %s FROM %s LIMIT 1", col, table))
	if err == nil {
		rows.Close() // CRITICAL: Close rows to prevent connection leak/hangs
		return       // Column exists
	}

	// If error, it implies column likely doesn't exist (or other DB error, but we attempt patch)
	fmt.Printf("[Auth] Auto-patching table %s: Adding column %s...\n", table, col)
	alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, def)
	if isMySQL {
		// MySQL syntax is slightly different but ADD COLUMN is standard
		alter = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, def)
	}
	_, err = db.Exec(alter)
	if err != nil {
		fmt.Printf("[Auth] Failed to patch column %s: %v\n", col, err)
	}
}

func getString(data map[string]interface{}, key, def string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return def
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
