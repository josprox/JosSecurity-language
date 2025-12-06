package core

import (
	"database/sql"
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

	// Get Prefix and Table Names
	prefix := "js_"
	if val, ok := r.Env["PREFIX"]; ok {
		prefix = val
	}
	storageTable := prefix + "storage"
	usersTable := prefix + "users"

	// Ensure DB tables exist
	r.ensureStorageTable(storageTable)

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

		// DB Registry
		if r.DB != nil {
			userId := r.getUserIdFromToken(usersTable, userToken)
			if userId > 0 {
				// Check if exists
				var existingId int
				check := fmt.Sprintf("SELECT id FROM %s WHERE user_id = ? AND path = ?", storageTable)
				err := r.DB.QueryRow(check, userId, fileName).Scan(&existingId)

				if err == sql.ErrNoRows {
					// Insert
					insert := fmt.Sprintf("INSERT INTO %s (user_id, path, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", storageTable)
					if val, ok := r.Env["DB"]; ok && val == "mysql" {
						insert = fmt.Sprintf("INSERT INTO %s (user_id, path, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", storageTable)
					}
					r.DB.Exec(insert, userId, fileName)
				} else {
					// Update timestamp
					update := fmt.Sprintf("UPDATE %s SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", storageTable)
					if val, ok := r.Env["DB"]; ok && val == "mysql" {
						update = fmt.Sprintf("UPDATE %s SET updated_at = NOW() WHERE id = ?", storageTable)
					}
					r.DB.Exec(update, existingId)
				}
			}
		}
		return true

	case "update":
		// UserStorage::update($token, $oldFile, $content, $newFile?)
		if len(args) < 3 {
			return false
		}
		userToken := fmt.Sprintf("%v", args[0])
		oldFileName := fmt.Sprintf("%v", args[1])
		content := fmt.Sprintf("%v", args[2])
		newFileName := oldFileName
		if len(args) >= 4 {
			newFileName = fmt.Sprintf("%v", args[3])
		}

		// Delete old file if name matches but content differs?
		// Actually update implies replacing content. Rename implies changing name.
		// If newFileName != oldFileName, we move/rename.

		baseDir := filepath.Join(basePath, userToken)
		oldPath := filepath.Join(baseDir, oldFileName)
		newPath := filepath.Join(baseDir, newFileName)

		// Ensure dir
		os.MkdirAll(filepath.Dir(newPath), 0755)

		// If renaming, remove old
		if oldFileName != newFileName {
			os.Remove(oldPath)
		}

		// Write new content
		if err := os.WriteFile(newPath, []byte(content), 0644); err != nil {
			return false
		}

		// Update DB
		if r.DB != nil {
			userId := r.getUserIdFromToken(usersTable, userToken)
			if userId > 0 {
				query := fmt.Sprintf("UPDATE %s SET path = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND path = ?", storageTable)
				if val, ok := r.Env["DB"]; ok && val == "mysql" {
					query = fmt.Sprintf("UPDATE %s SET path = ?, updated_at = NOW() WHERE user_id = ? AND path = ?", storageTable)
				}
				r.DB.Exec(query, newFileName, userId, oldFileName)
			}
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
			// Continue to delete from DB even if file missing
		}

		// DB Registry
		if r.DB != nil {
			userId := r.getUserIdFromToken(usersTable, userToken)
			if userId > 0 {
				query := fmt.Sprintf("DELETE FROM %s WHERE user_id = ? AND path = ?", storageTable)
				r.DB.Exec(query, userId, fileName)
			}
		}
		return true
	}
	return nil
}

// Helper to get User ID
func (r *Runtime) getUserIdFromToken(usersTable, token string) int {
	if r.DB == nil {
		return 0
	}
	var id int
	query := fmt.Sprintf("SELECT id FROM %s WHERE user_token = ? LIMIT 1", usersTable)
	err := r.DB.QueryRow(query, token).Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

var storageTableEnsured bool

func (r *Runtime) ensureStorageTable(tableName string) {
	if r.DB == nil || storageTableEnsured {
		return
	}

	createCtx := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		path VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`, tableName)

	if val, ok := r.Env["DB"]; ok && val == "mysql" {
		createCtx = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			path VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`, tableName)
	}

	r.DB.Exec(createCtx)
	storageTableEnsured = true
}
