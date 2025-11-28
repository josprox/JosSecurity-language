package core

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/parser"
)

// Cron Implementation (Daemon mode simulation)
func (r *Runtime) executeCronMethod(instance *Instance, method string, args []interface{}) interface{} {
	if method == "schedule" {
		if len(args) >= 3 {
			name := args[0].(string)
			timeStr := args[1].(string)

			if block, ok := args[2].(*parser.BlockStatement); ok {
				fmt.Printf("[Cron] Tarea '%s' programada para '%s'\n", name, timeStr)
				// Execute in goroutine
				go func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("[Cron] Error en tarea %s: %v\n", name, r)
						}
					}()
					r.executeBlock(block)
				}()
			}
		}
	}
	return nil
}
