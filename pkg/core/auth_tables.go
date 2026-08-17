package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var authTablesEnsured sync.Map

// ensureAuthTables creates and auto-migrates Auth tables using Schema & GranDB (100% DB-agnostic)
func (r *Runtime) ensureAuthTables(usersTable, rolesTable, prefix string) {
	db := r.GetDB()
	if db == nil {
		return
	}
	ensureKey := fmt.Sprintf("%p:%s", db, usersTable)
	if _, exists := authTablesEnsured.Load(ensureKey); exists {
		return
	}

	// 1. Tabla de Roles con Schema::create
	r.executeSchemaMethod(nil, "create", []interface{}{
		"roles",
		map[string]interface{}{
			"id":   "increments",
			"name": "string(50)|unique",
		},
	})

	// 2. Tabla de Usuarios con Schema::create
	r.executeSchemaMethod(nil, "create", []interface{}{
		"users",
		map[string]interface{}{
			"id":               "bigIncrements",
			"user_token":       "string(128)|nullable",
			"username":         "string(50)|nullable",
			"first_name":       "string(100)|nullable",
			"last_name":        "string(100)|nullable",
			"email":            "string(191)|unique",
			"phone":            "string(20)|nullable",
			"password":         "string(255)",
			"role_id":          "integer|default(2)",
			"verificado":       "integer|default(0)",
			"token_expires_at": "timestamp|nullable",
			"last_login_at":    "timestamp|nullable",
			"created_at":       "timestamp|nullable",
			"updated_at":       "timestamp|nullable",
		},
	})

	// 3. Tabla de Password Resets con Schema::create
	r.executeSchemaMethod(nil, "create", []interface{}{
		"password_resets",
		map[string]interface{}{
			"id":         "increments",
			"email":      "string(191)",
			"token":      "string(255)",
			"created_at": "timestamp|nullable",
			"expires_at": "timestamp|nullable",
			"used":       "integer|default(0)",
		},
	})

	// 4. Auto-Patching de columnas en tablas existentes usando Schema::hasColumn y Schema::table
	patchColumns := map[string]string{
		"username":         "string(50)|nullable",
		"user_token":       "string(128)|nullable",
		"first_name":       "string(100)|nullable",
		"last_name":        "string(100)|nullable",
		"phone":            "string(20)|nullable",
		"verificado":       "integer|default(0)",
		"token_expires_at": "timestamp|nullable",
		"last_login_at":    "timestamp|nullable",
		"created_at":       "timestamp|nullable",
		"updated_at":       "timestamp|nullable",
	}

	for colName, colType := range patchColumns {
		hasCol, _ := r.executeSchemaMethod(nil, "hasColumn", []interface{}{"users", colName}).(bool)
		if !hasCol {
			r.executeSchemaMethod(nil, "table", []interface{}{
				"users",
				map[string]interface{}{colName: colType},
			})
		}
	}

	// 5. Inserción de Roles por defecto usando GranDB
	for _, roleName := range []string{"admin", "client"} {
		var count int
		checkQuery := fmt.Sprintf("SELECT count(*) FROM %s WHERE name = ?", rolesTable)
		if err := db.QueryRow(checkQuery, roleName).Scan(&count); err == nil && count == 0 {
			r.insertFromMap(rolesTable, map[string]interface{}{
				"name": roleName,
			}, false)
		}
	}

	authTablesEnsured.Store(ensureKey, true)
}

func normalizeAuthEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func parseAuthExpiry(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed, true
		}
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func getString(data map[string]interface{}, key, def string) string {
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return def
}
