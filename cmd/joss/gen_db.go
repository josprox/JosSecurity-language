package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

var databaseIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateDatabaseIdentifier(name string) error {
	if !databaseIdentifierPattern.MatchString(name) {
		return fmt.Errorf("identificador de base de datos invalido: %q", name)
	}
	return nil
}

func getColumns(db *sql.DB, dbType, tableName string) ([]ColumnSchema, error) {
	if err := validateDatabaseIdentifier(tableName); err != nil {
		return nil, err
	}
	switch dbType {
	case "sqlite":
		return getSQLiteColumns(db, tableName)
	case "postgres", "postgresql", "pgx":
		return getPostgresColumns(db, tableName)
	default:
		return getMySQLColumns(db, tableName)
	}
}

func getSQLiteColumns(db *sql.DB, tableName string) ([]ColumnSchema, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnSchema
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, ColumnSchema{Name: name, Type: ctype})
		if err := validateDatabaseIdentifier(name); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func getPostgresColumns(db *sql.DB, tableName string) ([]ColumnSchema, error) {
	rows, err := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position", tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnSchema
	for rows.Next() {
		var field, columnType string
		if err := rows.Scan(&field, &columnType); err != nil {
			return nil, err
		}
		cols = append(cols, ColumnSchema{Name: field, Type: columnType})
		if err := validateDatabaseIdentifier(field); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func getMySQLColumns(db *sql.DB, tableName string) ([]ColumnSchema, error) {
	rows, err := db.Query(fmt.Sprintf("DESCRIBE `%s`", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnSchema
	for rows.Next() {
		var field, ctype, null, key string
		var extra interface{}
		var def interface{}
		if err := rows.Scan(&field, &ctype, &null, &key, &def, &extra); err != nil {
			return nil, err
		}
		cols = append(cols, ColumnSchema{Name: field, Type: ctype})
		if err := validateDatabaseIdentifier(field); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func getDisplayColumn(db *sql.DB, dbType, tableName string) (string, bool, error) {
	cols, err := getColumns(db, dbType, tableName)
	if err != nil {
		return "", false, err
	}
	if len(cols) == 0 {
		return "", false, nil
	}

	candidates := []string{"name", "title", "username", "email", "first_name", "last_name", "description", "slug", "code"}
	for _, candidate := range candidates {
		for _, col := range cols {
			if col.Name == candidate {
				return candidate, true, nil
			}
		}
	}

	for _, col := range cols {
		lowerType := strings.ToLower(col.Type)
		if strings.Contains(lowerType, "char") || strings.Contains(lowerType, "text") {
			return col.Name, true, nil
		}
	}

	return "id", true, nil
}
