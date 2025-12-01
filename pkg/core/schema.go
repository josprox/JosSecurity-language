package core

import (
	"fmt"
	"strings"

	"github.com/jossecurity/joss/pkg/parser"
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
			if !strings.HasPrefix(tableName, "js_") {
				tableName = "js_" + tableName
			}

			var definitions []string

			// Check if second argument is a function (closure-based approach)
			if fnLit, ok := args[1].(*parser.FunctionLiteral); ok {
				// Get the registered Blueprint class
				blueprintClass, ok := r.Classes["Blueprint"]
				if !ok {
					fmt.Println("[Schema] Error: Blueprint class not registered")
					return nil
				}

				// Create a Blueprint instance to collect column definitions
				blueprint := &Instance{
					Class:  blueprintClass,
					Fields: make(map[string]interface{}),
				}
				blueprint.Fields["_columns"] = []map[string]string{}

				fmt.Printf("[Schema] Created Blueprint instance, class: %s\n", blueprint.Class.Name.Value)

				// Call the function with the blueprint
				r.Variables["$table"] = blueprint
				if len(fnLit.Parameters) > 0 {
					r.Variables[fnLit.Parameters[0].Value] = blueprint
					fmt.Printf("[Schema] Set parameter %s to Blueprint instance\n", fnLit.Parameters[0].Value)
				}
				r.executeBlock(fnLit.Body)

				// Extract column definitions from blueprint
				if cols, ok := blueprint.Fields["_columns"].([]map[string]string); ok {
					for _, col := range cols {
						def := r.buildColumnDefinition(col["name"], col["type"], dbDriver)
						definitions = append(definitions, def)
					}
				}
			} else {
				// Map-based approach (legacy)
				colsMap, ok := args[1].(map[string]interface{})
				if !ok {
					fmt.Println("[Schema] Error: El segundo argumento debe ser un mapa de columnas.")
					return nil
				}

				for colName, colTypeRaw := range colsMap {
					colType := colTypeRaw.(string)
					def := r.buildColumnDefinition(colName, colType, dbDriver)
					definitions = append(definitions, def)
				}
			}

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
			if !strings.HasPrefix(tableName, "js_") {
				tableName = "js_" + tableName
			}
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
		return fmt.Sprintf("%s %s", name, sqlDef)
	case "string":
		sqlDef = "VARCHAR(255)"
	case "text":
		sqlDef = "TEXT"
	case "integer":
		sqlDef = "INT"
	case "decimal":
		sqlDef = "DECIMAL(10,2)"
	case "boolean":
		if driver == "sqlite" {
			sqlDef = "BOOLEAN"
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
			start := strings.Index(mod, "(")
			end := strings.LastIndex(mod, ")")
			if start != -1 && end != -1 {
				val := mod[start+1 : end]
				def += fmt.Sprintf(" DEFAULT %s", val)
			}
		}
	}

	// Add NOT NULL if not nullable and not primary key
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

// Blueprint method execution
func (r *Runtime) executeBlueprintMethod(instance *Instance, method string, args []interface{}) interface{} {
	fmt.Printf("[Blueprint] Method called: %s, args: %v\n", method, args)
	cols, _ := instance.Fields["_columns"].([]map[string]string)

	switch method {
	case "id":
		cols = append(cols, map[string]string{"name": "id", "type": "increments"})
	case "string":
		if len(args) > 0 {
			cols = append(cols, map[string]string{"name": args[0].(string), "type": "string"})
		}
	case "text":
		if len(args) > 0 {
			cols = append(cols, map[string]string{"name": args[0].(string), "type": "text"})
		}
	case "integer":
		if len(args) > 0 {
			cols = append(cols, map[string]string{"name": args[0].(string), "type": "integer"})
		}
	case "decimal":
		if len(args) > 0 {
			cols = append(cols, map[string]string{"name": args[0].(string), "type": "decimal"})
		}
	case "timestamps":
		cols = append(cols, map[string]string{"name": "created_at", "type": "datetime|default(CURRENT_TIMESTAMP)"})
		cols = append(cols, map[string]string{"name": "updated_at", "type": "datetime|default(CURRENT_TIMESTAMP)"})
	}

	instance.Fields["_columns"] = cols
	return nil
}
