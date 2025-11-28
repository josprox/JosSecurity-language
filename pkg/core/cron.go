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
			if r.DB != nil {
				// Check if exists
				var id int
				err := r.DB.QueryRow("SELECT id FROM js_cron WHERE name = ?", name).Scan(&id)
				if err == sql.ErrNoRows {
					_, err = r.DB.Exec("INSERT INTO js_cron (name, schedule, status) VALUES (?, ?, 'idle')", name, schedule)
				} else if err == nil {
					_, err = r.DB.Exec("UPDATE js_cron SET schedule = ? WHERE id = ?", schedule, id)
				}
				if err != nil {
					fmt.Printf("[Cron] Error registrando tarea %s: %v\n", name, err)
				}
			}

			if block, ok := args[2].(*parser.BlockStatement); ok {
				// 2. Check if we should run (Locking)
				if r.DB != nil {
					var isRunning bool
					err := r.DB.QueryRow("SELECT is_running FROM js_cron WHERE name = ?", name).Scan(&isRunning)
					if err == nil && isRunning {
						fmt.Printf("[Cron] Tarea '%s' ya está en ejecución. Saltando.\n", name)
						return nil
					}

					// Lock
					_, err = r.DB.Exec("UPDATE js_cron SET is_running = 1, status = 'running' WHERE name = ?", name)
					if err != nil {
						fmt.Printf("[Cron] Error bloqueando tarea %s: %v\n", name, err)
						return nil
					}
				}

				fmt.Printf("[Cron] Ejecutando tarea '%s'...\n", name)

				// Execute in goroutine
				go func() {
					defer func() {
						if rec := recover(); rec != nil {
							fmt.Printf("[Cron] Error en tarea %s: %v\n", name, rec)
							if r.DB != nil {
								r.DB.Exec("UPDATE js_cron SET is_running = 0, status = 'error', last_run_at = CURRENT_TIMESTAMP WHERE name = ?", name)
							}
						} else {
							if r.DB != nil {
								r.DB.Exec("UPDATE js_cron SET is_running = 0, status = 'completed', last_run_at = CURRENT_TIMESTAMP WHERE name = ?", name)
							}
						}
					}()
					r.executeBlock(block)
				}()
			}
		}
	}
	return nil
}
