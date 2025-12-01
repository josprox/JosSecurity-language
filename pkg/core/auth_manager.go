package core

import (
	"fmt"
)

// EnsureAuthTables creates js_roles and js_users if they don't exist
func (r *Runtime) EnsureAuthTables() {
	if r.DB == nil {
		return
	}

	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	rolesTable := prefix + "roles"
	usersTable := prefix + "users"

	dbDriver := "mysql"
	if val, ok := r.Env["DB"]; ok {
		dbDriver = val
	}

	// 1. Create Roles Table
	var queryRoles string
	if dbDriver == "sqlite" {
		queryRoles = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(50) UNIQUE
		)`, rolesTable)
	} else {
		queryRoles = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) UNIQUE
		)`, rolesTable)
	}
	r.DB.Exec(queryRoles)

	// Seed Roles
	// SQLite doesn't support INSERT IGNORE. Use INSERT OR IGNORE.
	var insertRole string
	if dbDriver == "sqlite" {
		insertRole = "INSERT OR IGNORE INTO"
	} else {
		insertRole = "INSERT IGNORE INTO"
	}
	r.DB.Exec(fmt.Sprintf("%s %s (id, name) VALUES (1, 'admin')", insertRole, rolesTable))
	r.DB.Exec(fmt.Sprintf("%s %s (id, name) VALUES (2, 'client')", insertRole, rolesTable))

	// 2. Create Users Table
	var queryUsers string
	if dbDriver == "sqlite" {
		queryUsers = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255),
			email VARCHAR(255) UNIQUE,
			password VARCHAR(255),
			role_id INTEGER DEFAULT 2,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (role_id) REFERENCES %s(id)
		)`, usersTable, rolesTable)
	} else {
		queryUsers = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255) UNIQUE,
			password VARCHAR(255),
			role_id INT DEFAULT 2,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (role_id) REFERENCES %s(id)
		)`, usersTable, rolesTable)
	}
	r.DB.Exec(queryUsers)
}
