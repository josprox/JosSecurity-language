package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var migrationNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func migrationTableName(name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if !migrationNamePattern.MatchString(normalized) {
		return "", fmt.Errorf("el nombre de migracion solo puede contener letras, numeros y guiones bajos")
	}
	normalized = strings.TrimPrefix(normalized, "create_")
	normalized = strings.TrimSuffix(normalized, "_table")
	if normalized == "" {
		return "", fmt.Errorf("el nombre de migracion no identifica una tabla")
	}
	return strings.ToLower(pluralize(singularize(normalized))), nil
}

func createMigration(name string) error {
	tableName, err := migrationTableName(name)
	if err != nil {
		return err
	}
	timestamp := time.Now().Format("20060102150405")
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	filename := fmt.Sprintf("%s_%s.joss", timestamp, normalizedName)
	path := filepath.Join("app", "database", "migrations", filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creando directorio de migraciones: %w", err)
	}

	content := fmt.Sprintf(`// Migration: %s
// Created at: %s

public class Create%sTable extends Migration {
    public func up() {
        // Schema::create automatically handles the prefix defined in env.joss
        Schema::create("%s", func(mixed $table) {
            $table->id()
            $table->string("name")
            $table->timestamps()
        })
    }

    public func down() {
        Schema::drop("%s")
    }
}
`, normalizedName, time.Now().Format("2006-01-02 15:04:05"), snakeToCamel(tableName), tableName, tableName)

	writeGenFile(path, content)
	return nil
}
