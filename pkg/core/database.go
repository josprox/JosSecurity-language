package core

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// Helper to get table prefix
func getTablePrefix() string {
	prefix := os.Getenv("DB_PREFIX")
	if prefix == "" {
		return "js_"
	}
	return prefix
}

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
			// Apply prefix if not already present (and not a raw query)
			prefix := getTablePrefix()
			if !strings.HasPrefix(tableName, prefix) {
				tableName = prefix + tableName
			}
			instance.Fields["_table"] = tableName
		}
		return instance // Return this for chaining

	case "select":
		if len(args) > 0 {
			if cols, ok := args[0].(string); ok {
				instance.Fields["_select"] = cols
			} else if cols, ok := args[0].([]interface{}); ok {
				// Handle array of columns
				strCols := []string{}
				for _, c := range cols {
					strCols = append(strCols, fmt.Sprintf("%v", c))
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
			table := instance.Fields["tabla"]
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
			return rowsToJSON(rows) // Default to JSON
		}

		// New fluent builder API
		wheres := instance.Fields["_wheres"].([]string)
		bindings := instance.Fields["_bindings"].([]interface{})

		if len(args) == 2 {
			col := args[0].(string)
			val := args[1]
			wheres = append(wheres, fmt.Sprintf("%s = ?", col))
			bindings = append(bindings, val)
		} else if len(args) == 3 {
			col := args[0].(string)
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
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("INNER JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "leftJoin":
		if len(args) >= 4 {
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
			if _, ok := instance.Fields["_joins"]; !ok {
				instance.Fields["_joins"] = []string{}
			}
			join := fmt.Sprintf("LEFT JOIN %s ON %s %s %s", table, first, op, second)
			instance.Fields["_joins"] = append(instance.Fields["_joins"].([]string), join)
		}
		return instance

	case "rightJoin":
		if len(args) >= 4 {
			table := args[0].(string)
			first := args[1].(string)
			op := args[2].(string)
			second := args[3].(string)
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
			return "[]"
		}

		table := instance.Fields["_table"].(string)
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
			return "[]"
		}
		defer rows.Close()

		return rowsToJSON(rows)

	case "first":
		if r.DB == nil {
			return nil
		}

		table := instance.Fields["_table"].(string)
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

		jsonStr := rowsToJSON(rows)
		if jsonStr == "[]" {
			return nil
		}
		return strings.TrimSuffix(strings.TrimPrefix(jsonStr.(string), "["), "]")

	case "insert":
		if r.DB == nil {
			return false
		}
		if len(args) > 0 {
			// Support insert(["col1", "col2"], ["val1", "val2"])
			if len(args) == 2 {
				cols := args[0].([]interface{})
				vals := args[1].([]interface{})

				if len(cols) != len(vals) {
					return false
				}

				table := instance.Fields["_table"].(string)
				colNames := []string{}
				placeholders := []string{}
				bindings := []interface{}{}

				for _, c := range cols {
					colNames = append(colNames, fmt.Sprintf("%v", c))
					placeholders = append(placeholders, "?")
				}
				for _, v := range vals {
					bindings = append(bindings, v)
				}

				query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(colNames, ", "), strings.Join(placeholders, ", "))

				_, err := r.DB.Exec(query, bindings...)
				if err != nil {
					fmt.Printf("[GranMySQL] Error insert: %v\n", err)
					return false
				}
				return true
			}
		}
		return false

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

func rowsToJSON(rows *sql.Rows) interface{} {
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
