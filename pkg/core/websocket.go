package core

import "fmt"

// WebSocket Implementation
func (r *Runtime) executeWebSocketMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "broadcast":
		if len(args) > 0 {
			msg := args[0]
			if BroadcastFunc != nil {
				BroadcastFunc(msg)
				return true
			} else {
				fmt.Println("[WebSocket] Error: BroadcastFunc not initialized")
			}
		}
		return false
	}
	return nil
}
