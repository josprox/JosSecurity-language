package core

import (
	"database/sql"
	"fmt"
	"strings"
)

// GranMySQL Implementation
func (r *Runtime) executeGranMySQLMethod(instance *Instance, method string, args []interface{}) interface{} {
	// Initialize internal state if needed
	if _, ok := instance.Fields["_wheres"]; !ok {
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_select"] = "*"
		instance.Fields["_table"] = ""
	}

	switch method {
	case "table":
		if len(args) > 0 {
			tableName := args[0].(string)
			instance.Fields["_table"] = quoteIdentifier(r.applyTablePrefix(tableName))
		}
		return instance // Return this for chaining

	case "select":
		if len(args) > 0 {
			if cols, ok := args[0].(string); ok {
				// Handle single string "table.col" or "col"
				instance.Fields["_select"] = cols
			} else if cols, ok := args[0].([]interface{}); ok {
				// Handle array of columns
				strCols := []string{}
				for _, c := range cols {
					colStr := fmt.Sprintf("%v", c)
					// Handle "table.col as alias"
					if strings.Contains(strings.ToLower(colStr), " as ") {
						parts := strings.Split(colStr, " ") // simplistic split
						// Try to find the part with "."
						for i, p := range parts {
							if strings.Contains(p, ".") {
								parts[i] = r.applyColumnPrefix(p)
							}
						}
						strCols = append(strCols, strings.Join(parts, " "))
					} else {
						strCols = append(strCols, r.applyColumnPrefix(colStr))
					}
				}
				instance.Fields["_select"] = strings.Join(strCols, ", ")
			}
		}
		return instance

	case "where":
		// Support both old and new API
		// Old API: where("json") - uses tabla, comparar, comparable properties
		// New API: where(col, val) or where(col, op, val) - fluent builder

		if len(args) == 1 {
			// Old API: where("json") or where("array")
			format := args[0].(string)

			// Use legacy properties
			// Use getTable helper
			table := r.getTable(instance)
			col := instance.Fields["comparar"]
			val := instance.Fields["comparable"]

			if r.DB == nil {
				return "[]"
			}

			query := fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", table, col)
			rows, err := r.DB.Query(query, val)
			if err != nil {
				fmt.Printf("[GranMySQL] Error en where: %v\n", err)
				return "[]"
			}
			defer rows.Close()

			// Return based on format
			if format == "json" {
				return rowsToJSON(rows)
			}
			return rowsToJSON(rows) // Default to JSON for legacy where()
		}

		// New fluent builder API
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		if len(args) == 2 {
			col := quoteIdentifier(r.applyColumnPrefix(args[0].(string)))
			val := args[1]
			wheres = append(wheres, fmt.Sprintf("%s = ?", col))
			bindings = append(bindings, val)
		} else if len(args) == 3 {
			col := quoteIdentifier(r.applyColumnPrefix(args[0].(string)))
			op := args[1].(string)
			val := args[2]
			wheres = append(wheres, fmt.Sprintf("%s %s ?", col, op))
			bindings = append(bindings, val)
		}

		instance.Fields["_wheres"] = wheres
		instance.Fields["_bindings"] = bindings
		return instance

	case "innerJoin":
		if len(args) >= 4 {
			table := r.applyTablePrefix(args[0].(string))
			first := r.applyColumnPrefix(args[1].(string))
			op := args[2].(string)
			second := r.applyColumnPrefix(args[3].(string))
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("INNER JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "leftJoin":
		if len(args) >= 4 {
			table := r.applyTablePrefix(args[0].(string))
			first := r.applyColumnPrefix(args[1].(string))
			op := args[2].(string)
			second := r.applyColumnPrefix(args[3].(string))
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("LEFT JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "rightJoin":
		if len(args) >= 4 {
			table := r.applyTablePrefix(args[0].(string))
			first := r.applyColumnPrefix(args[1].(string))
			op := args[2].(string)
			second := r.applyColumnPrefix(args[3].(string))
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("RIGHT JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "get":
		if r.DB == nil {
			fmt.Println("[GranMySQL] Error: No DB connection")
			return []map[string]interface{}{}
		}

		table := r.getTable(instance)
		sel := instance.Fields["_select"].(string)
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		query := fmt.Sprintf("SELECT %s FROM %s", sel, table)

		if joins, ok := instance.Fields["_joins"]; ok {
			for _, j := range joins.([]string) {
				query += " " + j
			}
		}

		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}

		// Reset state after query build
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_select"] = "*"
		instance.Fields["_joins"] = []string{}

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			fmt.Printf("[GranMySQL] Error en get: %v\n", err)
			return []map[string]interface{}{}
		}
		defer rows.Close()

		return rowsToMap(rows)

	case "first":
		if r.DB == nil {
			return nil
		}

		table := r.getTable(instance)
		sel := instance.Fields["_select"].(string)
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		query := fmt.Sprintf("SELECT %s FROM %s", sel, table)

		if joins, ok := instance.Fields["_joins"]; ok {
			for _, j := range joins.([]string) {
				query += " " + j
			}
		}

		if len(wheres) > 0 {
			query += " WHERE " + strings.Join(wheres, " AND ")
		}
		query += " LIMIT 1"

		// Reset state
		instance.Fields["_wheres"] = []string{}
		instance.Fields["_bindings"] = []interface{}{}
		instance.Fields["_joins"] = []string{}

		rows, err := r.DB.Query(query, bindings...)
		if err != nil {
			return nil
		}
		defer rows.Close()

		results := rowsToMap(rows)
		if len(results) > 0 {
			return results[0]
		}
		return nil

	case "insert":
		return r.executeInsertMethod(instance, args)

	case "update":
		return r.executeUpdateMethod(instance, args)

	case "delete":
		return r.executeDeleteMethod(instance)

	case "deleteAll":
		return r.executeDeleteAllMethod(instance)

	case "truncate":
		return r.executeTruncateMethod(instance)

	case "query":
		if len(args) > 0 {
			if sqlStr, ok := args[0].(string); ok {
				if r.DB == nil {
					return nil
				}
				_, err := r.DB.Exec(sqlStr)
				if err != nil {
					fmt.Printf("[GranMySQL] Error query: %v\n", err)
					return false
				}
				return true
			}
		}
	}
	return nil
}

func rowsToMap(rows *sql.Rows) []map[string]interface{} {
	var results []map[string]interface{}
	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range cols {
		valPtrs[i] = &vals[i]
	}

	for rows.Next() {
		rows.Scan(valPtrs...)
		row := make(map[string]interface{})
		for i, colName := range cols {
			valVal := vals[i]
			if b, ok := valVal.([]byte); ok {
				row[colName] = string(b)
			} else {
				row[colName] = valVal
			}
		}
		results = append(results, row)
	}
	return results
}

func rowsToJSON(rows *sql.Rows) string {
	var results []string
	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range cols {
		valPtrs[i] = &vals[i]
	}

	for rows.Next() {
		rows.Scan(valPtrs...)
		rowStr := "{"
		for i, colName := range cols {
			valVal := vals[i]
			if b, ok := valVal.([]byte); ok {
				valVal = string(b)
			}
			rowStr += fmt.Sprintf("\"%s\": \"%v\"", colName, valVal)
			if i < len(cols)-1 {
				rowStr += ", "
			}
		}
		rowStr += "}"
		results = append(results, rowStr)
	}
	return "[" + strings.Join(results, ", ") + "]"
}

// Helper to quote identifiers (basic protection)
func quoteIdentifier(name string) string {
	name = strings.TrimSpace(name)
	if name == "*" {
		return "*"
	}
	// Don't quote if it contains spaces (likely a function or complex expression)
	if strings.Contains(name, " ") || strings.Contains(name, "(") {
		return name
	}
	// Handle table.column
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, p := range parts {
			parts[i] = quoteIdentifier(p)
		}
		return strings.Join(parts, ".")
	}
	if strings.HasPrefix(name, "`") && strings.HasSuffix(name, "`") {
		return name
	}
	return "`" + name + "`"
}

// Helper to apply prefix to table names
func (r *Runtime) applyTablePrefix(name string) string {
	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	if prefix == "" {
		return name
	}
	if !strings.HasPrefix(name, prefix) {
		return prefix + name
	}
	return name
}

// Helper to apply prefix to column names (only if qualified with table)
func (r *Runtime) applyColumnPrefix(name string) string {
	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	if prefix == "" {
		return name
	}

	// Handle "table.column"
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		tablePart := parts[0]
		colPart := parts[1]

		// Don't prefix if already prefixed
		if !strings.HasPrefix(tablePart, prefix) {
			tablePart = prefix + tablePart
		}
		return tablePart + "." + colPart
	}

	// If no dot, return as is (it's a column name)
	return name
}

// Helper to get table name from instance
// Checks _table (internal), tabla (legacy property), or infers from class name
func (r *Runtime) getTable(instance *Instance) string {
	// 1. Check internal _table field (set via table() method)
	if val, ok := instance.Fields["_table"]; ok {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}

	// 2. Check public tabla property (set in constructor)
	if val, ok := instance.Fields["tabla"]; ok {
		if str, ok := val.(string); ok && str != "" {
			// Sync to _table for future use
			instance.Fields["_table"] = str
			return str
		}
	}

	// 3. Infer from Class Name (e.g. User -> js_users)
	className := instance.Class.Name.Value
	if className == "GranDB" || className == "Model" {
		return ""
	}

	// Simple pluralization and snake_case
	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	tableName := prefix + strings.ToLower(className) + "s"

	// Sync to _table
	instance.Fields["_table"] = tableName

	return tableName
}
