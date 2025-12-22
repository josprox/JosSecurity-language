package core

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// UserStorage Native Class Implementation
// Usage: UserStorage::put($user_token, "profile.jpg", $file_content)
//
//	UserStorage::path($user_token, "profile.jpg")
func (r *Runtime) executeUserStorageMethod(instance *Instance, method string, args []interface{}) interface{} {
	basePath := "assets/users"
	storageType := "local"
	if val, ok := r.Env["STORAGE"]; ok {
		storageType = val
	}

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

		// DB Registry Logic (Common for both)
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

		if storageType == "OCI" {
			return r.ociPut(userToken, fileName, content)
		} else {
			// LOCAL STORAGE
			fullPath := filepath.Join(basePath, userToken, fileName)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("[Storage DEBUG] MkdirAll error: %v\n", err)
				return false
			}

			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				fmt.Printf("[Storage DEBUG] WriteFile error: %v\n", err)
				return false
			}
			return true
		}

	case "get":
		if len(args) < 2 {
			return nil
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1])

		if storageType == "OCI" {
			return r.ociGet(userToken, fileName)
		} else {
			fullPath := filepath.Join(basePath, userToken, fileName)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return nil
			}
			return string(content)
		}

	case "delete":
		if len(args) < 2 {
			return false
		}
		userToken := fmt.Sprintf("%v", args[0])
		fileName := fmt.Sprintf("%v", args[1])

		// DB Registry Delete
		if r.DB != nil {
			userId := r.getUserIdFromToken(usersTable, userToken)
			if userId > 0 {
				query := fmt.Sprintf("DELETE FROM %s WHERE user_id = ? AND path = ?", storageTable)
				r.DB.Exec(query, userId, fileName)
			}
		}

		if storageType == "OCI" {
			return r.ociDelete(userToken, fileName)
		} else {
			fullPath := filepath.Join(basePath, userToken, fileName)
			if err := os.Remove(fullPath); err != nil {
				return false
			}
			return true
		}
	}
	return nil
}

// --- OCI Helpers ---

func (r *Runtime) getOCIClient() (objectstorage.ObjectStorageClient, context.Context, error) {
	privateKeyPath := r.Env["OCI_PRIVATE_KEY_PATH"]
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return objectstorage.ObjectStorageClient{}, nil, err
	}

	passphrase := r.Env["OCI_PASSPHRASE"]
	confProvider := common.NewRawConfigurationProvider(
		r.Env["OCI_TENANCY_ID"],
		r.Env["OCI_USER_ID"],
		r.Env["OCI_REGION"],
		r.Env["OCI_FINGERPRINT"],
		string(privateKey),
		&passphrase,
	)

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(confProvider)
	return client, context.Background(), err
}

func (r *Runtime) ociPut(userToken, fileName, content string) bool {
	fmt.Println("[OCI Debug] Initializing Client...")
	client, ctx, err := r.getOCIClient()
	if err != nil {
		fmt.Printf("[OCI Error] Client Init: %v\n", err)
		return false
	}

	// Add Timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	namespace := r.Env["OCI_NAMESPACE"]
	bucketName := r.Env["OCI_BUCKET_NAME"]
	objectName := userToken + "/" + fileName // Use forward slash for OCI

	req := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &objectName,
		PutObjectBody: io.NopCloser(strings.NewReader(content)),
		ContentLength: common.Int64(int64(len(content))),
	}

	fmt.Printf("[OCI Debug] Uploading %s (%d bytes)...\n", objectName, len(content))
	_, err = client.PutObject(ctx, req)
	if err != nil {
		fmt.Printf("[OCI Error] PutObject: %v\n", err)
		return false
	}
	fmt.Println("[OCI Debug] Upload Success!")
	return true
}

func (r *Runtime) ociGet(userToken, fileName string) interface{} {
	client, ctx, err := r.getOCIClient()
	if err != nil {
		return nil
	}

	namespace := r.Env["OCI_NAMESPACE"]
	bucketName := r.Env["OCI_BUCKET_NAME"]
	objectName := userToken + "/" + fileName

	req := objectstorage.GetObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &objectName,
	}

	resp, err := client.GetObject(ctx, req)
	if err != nil {
		return nil
	}
	defer resp.Content.Close()

	content, err := io.ReadAll(resp.Content)
	if err != nil {
		return nil
	}
	return string(content)
}

func (r *Runtime) ociDelete(userToken, fileName string) bool {
	client, ctx, err := r.getOCIClient()
	if err != nil {
		return false
	}

	namespace := r.Env["OCI_NAMESPACE"]
	bucketName := r.Env["OCI_BUCKET_NAME"]
	objectName := userToken + "/" + fileName

	req := objectstorage.DeleteObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
		ObjectName:    &objectName,
	}

	_, err = client.DeleteObject(ctx, req)
	return err == nil
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
