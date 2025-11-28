package core

import (
	"database/sql"
	"fmt"
)

// EnsureMigrationTable creates the js_migration table if it doesn't exist
func (r *Runtime) EnsureMigrationTable() {
	if r.DB == nil {
		return
	}

	query := `
	CREATE TABLE IF NOT EXISTS js_migration (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		migration VARCHAR(255) NOT NULL,
		batch INTEGER NOT NULL,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	dbDriver := "mysql"
	if val, ok := r.Env["DB"]; ok {
		dbDriver = val
	}

	if dbDriver == "mysql" {
		query = `
		CREATE TABLE IF NOT EXISTS js_migration (
			id INT AUTO_INCREMENT PRIMARY KEY,
			migration VARCHAR(255) NOT NULL,
			batch INT NOT NULL,
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`
	}

	_, err := r.DB.Exec(query)
	if err != nil {
		fmt.Printf("[Migration] Error creando tabla js_migration: %v\n", err)
	}
}

// GetExecutedMigrations returns a map of executed migration filenames
func (r *Runtime) GetExecutedMigrations() map[string]bool {
	executed := make(map[string]bool)
	if r.DB == nil {
		return executed
	}

	rows, err := r.DB.Query("SELECT migration FROM js_migration")
	if err != nil {
		return executed
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			executed[name] = true
		}
	}
	return executed
}

// GetNextBatch returns the next batch number
func (r *Runtime) GetNextBatch() int {
	if r.DB == nil {
		return 1
	}
	var maxBatch sql.NullInt64
	err := r.DB.QueryRow("SELECT MAX(batch) FROM js_migration").Scan(&maxBatch)
	if err != nil {
		return 1
	}
	if maxBatch.Valid {
		return int(maxBatch.Int64) + 1
	}
	return 1
}

// LogMigration logs a successful migration
func (r *Runtime) LogMigration(migration string, batch int) {
	if r.DB == nil {
		return
	}
	_, err := r.DB.Exec("INSERT INTO js_migration (migration, batch) VALUES (?, ?)", migration, batch)
	if err != nil {
		fmt.Printf("[Migration] Error registrando migración %s: %v\n", migration, err)
	}
}
