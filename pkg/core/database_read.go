package core

import (
	"fmt"
	"strings"
)

// executeGetMethod handles .get()
func (r *Runtime) executeGetMethod(instance *Instance, args []interface{}) interface{} {
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

	// Add Order By
	if order, ok := instance.Fields["_order"]; ok {
		query += " ORDER BY " + order.(string)
	}

	// Add Limit
	if limit, ok := instance.Fields["_limit"]; ok {
		query += fmt.Sprintf(" LIMIT %d", limit.(int))
	}

	// Add Offset
	if offset, ok := instance.Fields["_offset"]; ok {
		query += fmt.Sprintf(" OFFSET %d", offset.(int))
	}

	// Reset state
	instance.Fields["_wheres"] = []string{}
	instance.Fields["_bindings"] = []interface{}{}
	instance.Fields["_select"] = "*"
	instance.Fields["_joins"] = []string{}
	delete(instance.Fields, "_order")
	delete(instance.Fields, "_limit")
	delete(instance.Fields, "_offset")

	rows, err := r.DB.Query(query, bindings...)
	if err != nil {
		fmt.Printf("[GranMySQL] Error en get: %v\n", err)
		return []map[string]interface{}{}
	}
	defer rows.Close()

	return rowsToMap(rows)
}

// executeFirstMethod handles .first()
func (r *Runtime) executeFirstMethod(instance *Instance, args []interface{}) interface{} {
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

	if order, ok := instance.Fields["_order"]; ok {
		query += " ORDER BY " + order.(string)
	}

	query += " LIMIT 1"

	// Reset state
	instance.Fields["_wheres"] = []string{}
	instance.Fields["_bindings"] = []interface{}{}
	instance.Fields["_joins"] = []string{}
	delete(instance.Fields, "_order")

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
}

// executeCountMethod handles .count()
func (r *Runtime) executeCountMethod(instance *Instance, args []interface{}) interface{} {
	if r.DB == nil {
		return 0
	}
	table := r.getTable(instance)
	wheres := instance.Fields["_wheres"].([]string)
	bindings := instance.Fields["_bindings"].([]interface{})

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)

	if joins, ok := instance.Fields["_joins"]; ok {
		for _, j := range joins.([]string) {
			query += " " + j
		}
	}

	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}

	instance.Fields["_wheres"] = []string{}
	instance.Fields["_bindings"] = []interface{}{}
	instance.Fields["_joins"] = []string{}

	var count int
	err := r.DB.QueryRow(query, bindings...).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
