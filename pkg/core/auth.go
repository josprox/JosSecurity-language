package core

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	// Ensure tables exist (lazy check removed, moved to explicit init)
	// But we can keep it here as a fallback or just rely on EnsureAuthTables being called.
	// To be safe and follow the pattern, let's move the logic to EnsureAuthTables and call it from Runtime.

	switch method {
	case "create":
		// Auth::create([email, password, name, role_id?])
		if len(args) > 0 {
			if data, ok := args[0].([]interface{}); ok && len(data) >= 2 {
				email := data[0].(string)
				password := data[1].(string)
				name := "User"
				if len(data) > 2 {
					name = data[2].(string)
				}
				roleId := 2 // Default to Client
				if len(data) > 3 {
					if rId, ok := data[3].(int64); ok {
						roleId = int(rId)
					} else if rId, ok := data[3].(int); ok {
						roleId = rId
					}
				}

				hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return false
				}
				hashedPassword := string(hashedBytes)

				if r.DB == nil {
					return false
				}

				nowFunc := "NOW()"
				if val, ok := r.Env["DB"]; ok && val == "sqlite" {
					nowFunc = "CURRENT_TIMESTAMP"
				}

				query := fmt.Sprintf("INSERT INTO %s (name, email, password, role_id, created_at, updated_at) VALUES (?, ?, ?, ?, %s, %s)", usersTable, nowFunc, nowFunc)
				_, err = r.DB.Exec(query, name, email, hashedPassword, roleId)
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
			var roleName string
			// Join with roles table to get role name
			query := fmt.Sprintf(`
				SELECT u.id, u.name, u.password, r.name 
				FROM %s u 
				LEFT JOIN %s r ON u.role_id = r.id 
				WHERE u.email = ?`, usersTable, rolesTable)

			err := r.DB.QueryRow(query, email).Scan(&userId, &userName, &storedHash, &roleName)
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

			// Store in session (for stateful support)
			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					sessInst.Fields["user_id"] = userId
					sessInst.Fields["user_name"] = userName
					sessInst.Fields["user_email"] = email
					sessInst.Fields["user_role"] = roleName
				}
			}

			// Return JWT Token (Compliance with Bible)
			return r.generateJWT(userId, email, userName, false)
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
	case "guest":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if _, ok := sessInst.Fields["user_id"]; ok {
					return false
				}
			}
		}
		return true
	case "user":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if id, ok := sessInst.Fields["user_id"]; ok {
					// Fetch full user from DB
					if r.DB != nil {
						var userId int
						var name, email, password, createdAt, updatedAt string
						var roleId int

						// Handle sqlite vs mysql dates if needed, but for now scan as string
						query := fmt.Sprintf("SELECT id, name, email, password, role_id, created_at, updated_at FROM %s WHERE id = ?", usersTable)
						err := r.DB.QueryRow(query, id).Scan(&userId, &name, &email, &password, &roleId, &createdAt, &updatedAt)
						if err == nil {
							return map[string]interface{}{
								"id":         userId,
								"name":       name,
								"email":      email,
								"role_id":    roleId,
								"created_at": createdAt,
								"updated_at": updatedAt,
							}
						}
					}

					// Fallback to session data if DB fails or not available
					userMap := make(map[string]interface{})
					userMap["id"] = id
					if name, ok := sessInst.Fields["user_name"]; ok {
						userMap["name"] = name
					}
					if email, ok := sessInst.Fields["user_email"]; ok {
						userMap["email"] = email
					}
					return userMap
				}
			}
		}
		return nil
	case "id":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				if id, ok := sessInst.Fields["user_id"]; ok {
					return id
				}
			}
		}
		return nil
	case "hasRole":
		if len(args) > 0 {
			requiredRole := args[0].(string)
			if sessVal, ok := r.Variables["$__session"]; ok {
				if sessInst, ok := sessVal.(*Instance); ok {
					if role, ok := sessInst.Fields["user_role"]; ok {
						return role == requiredRole
					}
				}
			}
		}
		return false
	case "logout":
		if sessVal, ok := r.Variables["$__session"]; ok {
			if sessInst, ok := sessVal.(*Instance); ok {
				delete(sessInst.Fields, "user_id")
				delete(sessInst.Fields, "user_name")
				delete(sessInst.Fields, "user_email")
				delete(sessInst.Fields, "user_role")
			}
		}
		return true
	case "refresh":
		return false
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
