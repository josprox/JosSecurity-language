package core

import (
	"fmt"
	"strings"
)

// Schema Implementation
func (r *Runtime) executeSchemaMethod(instance *Instance, method string, args []interface{}) interface{} {
	if r.DB == nil {
		return nil
	}

	dbDriver := "mysql"
	if val, ok := r.Env["DB"]; ok {
		dbDriver = val
	}

	switch method {
	case "create":
		if len(args) >= 2 {
			tableName := args[0].(string)
			// colsMap := args[1].(map[string]interface{})
			// The evaluator evaluates arguments before calling.
			// So args[1] should be a map[string]interface{} or similar if we support map literals.
			// But wait, `parseBraceExpression` creates `MapLiteral`.
			// `Evaluate` converts `MapLiteral` to... what?
			// I need to check `evaluator.go`.
			// Assuming it evaluates to map[string]interface{} or similar.
			// Let's assume args[1] is map[string]interface{}.

			colsMap, ok := args[1].(map[string]interface{})
			if !ok {
				// It might be that map literal evaluation isn't fully supported to Go map yet?
				// Or maybe it's passed as something else.
				// Let's assume for now we can iterate it.
				fmt.Println("[Schema] Error: El segundo argumento debe ser un mapa de columnas.")
				return nil
			}

			var definitions []string

			// We need order? Maps are unordered.
			// If order matters (it usually does for readability, but not strictly for SQL), we might have issues.
			// But for `create table`, order is preserved in the DB schema usually.
			// If the user provides a map, they accept unordered unless we use an array of maps or ordered map.
			// For now, map is fine.

			// Handle "id": "increments" specially to put it first if possible?
			// Or just iterate.

			for colName, colTypeRaw := range colsMap {
				colType := colTypeRaw.(string)
				def := r.buildColumnDefinition(colName, colType, dbDriver)
				definitions = append(definitions, def)
			}

			// Simple primary key handling if not in definitions
			// (buildColumnDefinition handles it)

			query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(definitions, ", "))

			fmt.Printf("[Schema] Ejecutando: %s\n", query)
			_, err := r.DB.Exec(query)
			if err != nil {
				fmt.Printf("[Schema] Error creando tabla %s: %v\n", tableName, err)
			}
		}

	case "drop":
		if len(args) >= 1 {
			tableName := args[0].(string)
			query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
			_, err := r.DB.Exec(query)
			if err != nil {
				fmt.Printf("[Schema] Error eliminando tabla %s: %v\n", tableName, err)
			}
		}
	}
	return nil
}

func (r *Runtime) buildColumnDefinition(name, typeStr, driver string) string {
	parts := strings.Split(typeStr, "|")
	baseType := parts[0]
	modifiers := parts[1:]

	var sqlDef string

	switch baseType {
	case "increments":
		if driver == "sqlite" {
			sqlDef = "INTEGER PRIMARY KEY AUTOINCREMENT"
		} else {
			sqlDef = "INT AUTO_INCREMENT PRIMARY KEY"
		}
		return fmt.Sprintf("%s %s", name, sqlDef) // Name is included? Yes.
	case "string":
		sqlDef = "VARCHAR(255)"
	case "text":
		sqlDef = "TEXT"
	case "integer":
		sqlDef = "INT"
	case "boolean":
		if driver == "sqlite" {
			sqlDef = "BOOLEAN" // SQLite uses 0/1 but accepts BOOLEAN
		} else {
			sqlDef = "TINYINT(1)"
		}
	case "datetime":
		sqlDef = "DATETIME"
	default:
		sqlDef = "VARCHAR(255)"
	}

	def := fmt.Sprintf("%s %s", name, sqlDef)

	for _, mod := range modifiers {
		if mod == "unique" {
			def += " UNIQUE"
		} else if mod == "nullable" {
			def += " NULL"
		} else if strings.HasPrefix(mod, "default") {
			// Extract value default('value')
			// Simple parsing
			start := strings.Index(mod, "(")
			end := strings.LastIndex(mod, ")")
			if start != -1 && end != -1 {
				val := mod[start+1 : end]
				def += fmt.Sprintf(" DEFAULT %s", val)
			}
		}
	}

	// If not nullable and not primary key, add NOT NULL?
	// Usually yes, unless "nullable" is specified.
	isNullable := false
	for _, mod := range modifiers {
		if mod == "nullable" {
			isNullable = true
		}
	}
	if !isNullable && baseType != "increments" {
		def += " NOT NULL"
	}

	return def
}
