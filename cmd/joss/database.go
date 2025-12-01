package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jossecurity/joss/pkg/core"
	_ "modernc.org/sqlite"
)

func changeDatabaseEngine(target string) {
	fmt.Printf("Cambiando motor de base de datos a: %s\n", target)

	// 1. Read current env
	envMap := readEnvFile("env.joss")
	currentDB := envMap["DB"]
	if currentDB == "" {
		currentDB = "mysql" // Default
	}

	if currentDB == target {
		fmt.Println("El motor de base de datos ya es " + target)
		return
	}

	// 2. Connect to Source
	srcDB, err := connectToDB(currentDB, envMap)
	if err != nil {
		fmt.Printf("Error conectando a origen (%s): %v\n", currentDB, err)
		return
	}
	defer srcDB.Close()

	// 3. Connect to Dest
	destDB, err := connectToDB(target, envMap)
	if err != nil {
		fmt.Printf("Error conectando a destino (%s): %v\n", target, err)
		return
	}
	defer destDB.Close()

	fmt.Println("Conectado a origen y destino.")

	// 3.5 Run Migrations on Destination to ensure Schema exists
	fmt.Println("Preparando esquema en base de datos destino...")
	destRt := core.NewRuntime()
	destRt.DB = destDB
	destRt.Env = make(map[string]string)
	// Copy env
	for k, v := range envMap {
		destRt.Env[k] = v
	}
	destRt.Env["DB"] = target // Force target driver

	// Ensure System Tables
	destRt.EnsureMigrationTable()
	destRt.EnsureAuthTables()
	destRt.EnsureCronTable()

	// Run User Migrations
	performMigrations(destRt)

	fmt.Println("Iniciando migración de datos...")

	// 4. Get Tables from Source
	tables, err := getTables(srcDB, currentDB)
	if err != nil {
		fmt.Printf("Error obteniendo tablas: %v\n", err)
		return
	}

	// 5. Migrate Data
	for _, table := range tables {
		prefix := "js_"
		if val, ok := envMap["PREFIX"]; ok {
			prefix = val
		}

		if table == "sqlite_sequence" || table == prefix+"migration" || table == prefix+"cron" {
			continue
		}

		fmt.Printf("Migrando tabla: %s... ", table)

		// Read data
		rows, err := srcDB.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			fmt.Printf("Error leyendo tabla %s: %v\n", table, err)
			continue
		}

		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range cols {
			valPtrs[i] = &vals[i]
		}

		// Prepare insert in dest
		count := 0
		placeholders := make([]string, len(cols))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		insertCmd := "INSERT INTO"
		if target == "mysql" {
			insertCmd = "INSERT IGNORE INTO"
		} else if target == "sqlite" {
			insertCmd = "INSERT OR IGNORE INTO"
		}
		query := fmt.Sprintf("%s %s (%s) VALUES (%s)", insertCmd, table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

		tx, _ := destDB.Begin()
		stmt, err := tx.Prepare(query)
		if err != nil {
			fmt.Printf("Error preparando insert (¿Existe la tabla?): %v\n", err)
			tx.Rollback()
			continue
		}

		for rows.Next() {
			rows.Scan(valPtrs...)
			_, err = stmt.Exec(vals...)
			if err != nil {
				fmt.Printf("Error insertando fila: %v\n", err)
			} else {
				count++
			}
		}
		stmt.Close()
		tx.Commit()
		rows.Close()
		fmt.Printf("OK (%d filas)\n", count)
	}

	// 6. Update env.joss
	updateEnvFile("env.joss", "DB", target)
	fmt.Println("Migración completada. Archivo env.joss actualizado.")
}

func connectToDB(driver string, env map[string]string) (*sql.DB, error) {
	if driver == "sqlite" {
		path := "database.sqlite"
		if p, ok := env["DB_PATH"]; ok {
			path = strings.Trim(p, "\"")
			path = strings.Trim(path, "'")
		}
		return sql.Open("sqlite", path)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", env["DB_USER"], env["DB_PASS"], env["DB_HOST"], env["DB_NAME"])
		return sql.Open("mysql", dsn)
	}
}

func getTables(db *sql.DB, driver string) ([]string, error) {
	var tables []string
	var query string
	if driver == "sqlite" {
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	} else {
		query = "SHOW TABLES"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func changeDatabasePrefix(newPrefix string) {
	fmt.Printf("Cambiando prefijo de base de datos a: %s\n", newPrefix)

	// 1. Read current env
	envMap := readEnvFile("env.joss")
	currentPrefix := envMap["PREFIX"]
	if currentPrefix == "" {
		currentPrefix = "js_" // Default
	}

	if currentPrefix == newPrefix {
		fmt.Println("El prefijo ya es " + newPrefix)
		return
	}

	// 2. Connect to DB
	dbDriver := envMap["DB"]
	if dbDriver == "" {
		dbDriver = "mysql" // Default
	}

	db, err := connectToDB(dbDriver, envMap)
	if err != nil {
		fmt.Printf("Error conectando a DB: %v\n", err)
		return
	}
	defer db.Close()

	// 3. Get Tables
	tables, err := getTables(db, dbDriver)
	if err != nil {
		fmt.Printf("Error obteniendo tablas: %v\n", err)
		return
	}

	// 4. Rename Tables
	count := 0
	for _, table := range tables {
		if strings.HasPrefix(table, currentPrefix) {
			newTableName := strings.Replace(table, currentPrefix, newPrefix, 1)
			fmt.Printf("Renombrando %s a %s... ", table, newTableName)

			var query string
			if dbDriver == "sqlite" {
				query = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", table, newTableName)
			} else {
				query = fmt.Sprintf("RENAME TABLE %s TO %s", table, newTableName)
			}

			_, err := db.Exec(query)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
				count++
			}
		}
	}

	// 5. Update Source Code (Models and Migrations)
	fmt.Println("Actualizando código fuente (Modelos y Migraciones)...")
	updateSourceCodePrefix(currentPrefix, newPrefix)

	// 6. Update env.joss
	updateEnvFile("env.joss", "PREFIX", newPrefix)
	fmt.Printf("Prefijo actualizado. %d tablas renombradas.\n", count)
}

func updateSourceCodePrefix(oldPrefix, newPrefix string) {
	dirs := []string{"app/models", "app/database/migrations"}
	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".joss") {
				content, err := ioutil.ReadFile(path)
				if err == nil {
					strContent := string(content)
					// Replace "oldPrefix" with "newPrefix"
					// We look for the prefix followed by a letter, inside quotes ideally, but simple replacement is safer for now given the context
					// To be safer, we can look for specific patterns like:
					// "js_users" -> "comion_users"
					// We should simply replace all occurrences of the old prefix if it looks like a table name start
					// But simple string replacement of oldPrefix -> newPrefix might be too aggressive if oldPrefix is common (e.g. "a_")
					// However, usually prefixes are unique enough (js_, comion_db).
					// Let's replace occurrences of `"` + oldPrefix and `'` + oldPrefix

					newContent := strings.ReplaceAll(strContent, "\""+oldPrefix, "\""+newPrefix)
					newContent = strings.ReplaceAll(newContent, "'"+oldPrefix, "'"+newPrefix)

					if strContent != newContent {
						err = ioutil.WriteFile(path, []byte(newContent), 0644)
						if err == nil {
							fmt.Printf("Actualizado: %s\n", path)
						} else {
							fmt.Printf("Error actualizando %s: %v\n", path, err)
						}
					}
				}
			}
			return nil
		})
	}
}
