package core

import (
	"github.com/google/uuid"
)

// UUID Native Class Implementation
func (r *Runtime) executeUUIDMethod(instance *Instance, method string, args []interface{}) interface{} {
	switch method {
	case "generate", "v4":
		return uuid.New().String()
	}
	return nil
}
