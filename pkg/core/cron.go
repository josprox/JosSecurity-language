package core

import (
	"database/sql"
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

// Cron Implementation (Daemon mode simulation)
func (r *Runtime) executeCronMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "schedule" {
		if len(args) >= 3 {
			name := args[0].(string)
			schedule := args[1].(string)

			// 1. Register/Update Task in DB
			if r.GetDB() != nil {
				prefix := "js_"
				if val, ok := r.Env["PREFIX"]; ok {
					prefix = val
				}
				tableName := prefix + "cron"

				// Check if exists
				var id int
				err := r.GetDB().QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE name = ?", tableName), name).Scan(&id)
				if err == sql.ErrNoRows {
					_, err = r.GetDB().Exec(fmt.Sprintf("INSERT INTO %s (name, schedule, status) VALUES (?, ?, 'idle')", tableName), name, schedule)
				} else if err == nil {
					_, err = r.GetDB().Exec(fmt.Sprintf("UPDATE %s SET schedule = ? WHERE id = ?", tableName), schedule, id)
				}
				if err != nil {
					fmt.Printf("[Cron] Error registrando tarea %s: %v\n", name, err)
				}
			}

			if block, ok := args[2].(*parser.BlockStatement); ok {
				// 2. Check if we should run (Locking)
				if r.GetDB() != nil {
					prefix := "js_"
					if val, ok := r.Env["PREFIX"]; ok {
						prefix = val
					}
					tableName := prefix + "cron"

					var isRunning bool
					err := r.GetDB().QueryRow(fmt.Sprintf("SELECT is_running FROM %s WHERE name = ?", tableName), name).Scan(&isRunning)
					if err == nil && isRunning {
						fmt.Printf("[Cron] Tarea '%s' ya está en ejecución. Saltando.\n", name)
						return nil
					}

					// Lock
					_, err = r.GetDB().Exec(fmt.Sprintf("UPDATE %s SET is_running = 1, status = 'running' WHERE name = ?", tableName), name)
					if err != nil {
						fmt.Printf("[Cron] Error bloqueando tarea %s: %v\n", name, err)
						return nil
					}
				}

				fmt.Printf("[Cron] Ejecutando tarea '%s'...\n", name)

				// Execute in goroutine
				newR := r.Fork()
				go func() {
					defer func() {
						prefix := "js_"
						if val, ok := r.Env["PREFIX"]; ok {
							prefix = val
						}
						tableName := prefix + "cron"

						if rec := recover(); rec != nil {
							fmt.Printf("[Cron] Error en tarea %s: %v\n", name, rec)
							if r.GetDB() != nil {
								r.GetDB().Exec(fmt.Sprintf("UPDATE %s SET is_running = 0, status = 'error', last_run_at = CURRENT_TIMESTAMP WHERE name = ?", tableName), name)
							}
						} else {
							if r.GetDB() != nil {
								r.GetDB().Exec(fmt.Sprintf("UPDATE %s SET is_running = 0, status = 'completed', last_run_at = CURRENT_TIMESTAMP WHERE name = ?", tableName), name)
							}
						}
					}()
					newR.executeBlock(block)
				}()
			}
		}
	}
	return nil
}
