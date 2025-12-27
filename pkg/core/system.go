package core

import (
	"fmt"
	"os/exec"
)

// System Implementation
func (r *Runtime) executeSystemMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "env":
		if len(args) > 0 {
			key := args[0].(string)
			if val, ok := r.Env[key]; ok {
				return val
			}
			if len(args) > 1 {
				return args[1] // Default value
			}
			return ""
		}
	case "Run":
		// Security Check
		allow, ok := r.Env["ALLOW_SYSTEM_RUN"]
		if !ok || (allow != "true" && allow != "1") {
			fmt.Println("[System::Security] Error: Ejecución de comandos bloqueada. Configure ALLOW_SYSTEM_RUN=true en su entorno.")
			return ""
		}

		if len(args) > 0 {
			cmdName := args[0].(string)
			cmdArgs := []string{}
			if len(args) > 1 {
				if list, ok := args[1].([]interface{}); ok {
					for _, arg := range list {
						cmdArgs = append(cmdArgs, fmt.Sprintf("%v", arg))
					}
				}
			}

			cmd := exec.Command(cmdName, cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("[System] Error ejecutando '%s': %v\n", cmdName, err)
				return ""
			}
			return string(output)
		}
	case "load_driver":
		if len(args) > 0 {
			path := args[0].(string)
			fmt.Printf("[System] Cargando driver externo desde: %s (Simulación)\n", path)
			return true
		}
	case "log":
		if len(args) > 0 {
			msg := fmt.Sprintf("%v", args[0])
			fmt.Println("[System Log] " + msg)
			return nil
		}
	}
	return nil
}
