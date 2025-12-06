package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// UserStorage Native Class Implementation
// Usage: UserStorage::put($user_token, "profile.jpg", $file_content)
//
//	UserStorage::path($user_token, "profile.jpg")
func (r *Runtime) executeUserStorageMethod(instance *Instance, method string, args []interface{}) interface{} {
	basePath := "assets/users"

	switch method {
	case "put":
		if len(args) < 3 {
			return false
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1]) // Can be "photos/my_pic.jpg"
		content := fmt.Sprintf("%v", args[2])

		// Full path: assets/users/{token}/{fileName}
		fullPath := filepath.Join(basePath, userToken, fileName)

		// Ensure the specific directory for this file exists
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return false
		}
		return true

	case "path":
		if len(args) < 2 {
			return nil
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1])
		return filepath.Join(basePath, userToken, fileName)

	case "exists":
		if len(args) < 2 {
			return false
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1])
		fullPath := filepath.Join(basePath, userToken, fileName)
		if _, err := os.Stat(fullPath); err == nil {
			return true
		}
		return false

	case "delete":
		if len(args) < 2 {
			return false
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1])
		fullPath := filepath.Join(basePath, userToken, fileName)
		if err := os.Remove(fullPath); err != nil {
			return false
		}
		return true
	}
	return nil
}
