package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

func handleUserStorage(provider string) {
	fmt.Printf("Configurando UserStorage para proveedor: %s\n", provider)

	switch strings.ToLower(provider) {
	case "local":
		configureLocal()
	case "oci":
		configureOCI()
	case "aws", "azure":
		fmt.Printf("El proveedor '%s' aún no está soportado.\n", provider)
	default:
		fmt.Printf("Proveedor desconocido: %s\n", provider)
	}
}

func configureLocal() {
	// 1. Update Env to STORAGE=local
	updateEnvVariable("STORAGE", "local")
	fmt.Println("Almacenamiento configurado a LOCAL.")

	// 2. Ask if user wants to migrate FROM OCI to Local
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("¿Deseas descargar los archivos desde OCI hacia local? (s/n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))

	if text == "s" {
		migrateFromOCI()
	}
}

func configureOCI() {
	reader := bufio.NewReader(os.Stdin)

	// Gather OCI Details
	config := make(map[string]string)

	fmt.Println("\n--- Configuración de Oracle Cloud Infrastructure (OCI) ---")

	fields := []struct {
		Key   string
		Label string
	}{
		{"OCI_NAMESPACE", "Namespace"},
		{"OCI_BUCKET_NAME", "Bucket Name"},
		{"OCI_TENANCY_ID", "Tenancy OCID"},
		{"OCI_USER_ID", "User OCID"},
		{"OCI_REGION", "Region (e.g. mx-queretaro-1)"},
		{"OCI_FINGERPRINT", "Fingerprint"},
		{"OCI_PRIVATE_KEY_PATH", "Private Key Path (e.g. storage/oci_api_key.pem)"},
		{"OCI_PASSPHRASE", "Passphrase (dejar vacío si no tiene)"},
	}

	for _, field := range fields {
		fmt.Printf("%s: ", field.Label)
		val, _ := reader.ReadString('\n')
		val = strings.TrimSpace(val)
		if val != "" {
			config[field.Key] = val
			updateEnvVariable(field.Key, val)
		}
	}

	updateEnvVariable("STORAGE", "OCI")
	fmt.Println("\nConfiguración OCI guardada.")

	// Ask for migration
	fmt.Print("¿Deseas subir los archivos locales existentes a OCI? (s/n): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))

	if text == "s" {
		migrateToOCI()
	}
}

func updateEnvVariable(key, value string) {
	envPath := "env.joss"
	content, err := os.ReadFile(envPath)
	if err != nil {
		fmt.Printf("Error leyendo env.joss: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	found := false
	newLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		// If not found, append it (maybe under a generic comment or at end)
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	output := strings.Join(newLines, "\n")
	err = os.WriteFile(envPath, []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error actualizando env.joss: %v\n", err)
	}
}

// --- Migration Logis ---

func getOCIClient() (objectstorage.ObjectStorageClient, context.Context, error) {
	// Read Env again to be sure
	envMap := loadEnvMap()

	// Create configuration provider
	// We need to support reading the private key from file
	privateKey, err := os.ReadFile(envMap["OCI_PRIVATE_KEY_PATH"])
	if err != nil {
		return objectstorage.ObjectStorageClient{}, nil, fmt.Errorf("error leyendo private key: %v", err)
	}

	passphrase := envMap["OCI_PASSPHRASE"]
	confProvider := common.NewRawConfigurationProvider(
		envMap["OCI_TENANCY_ID"],
		envMap["OCI_USER_ID"],
		envMap["OCI_REGION"],
		envMap["OCI_FINGERPRINT"],
		string(privateKey),
		&passphrase,
	)

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(confProvider)
	if err != nil {
		return objectstorage.ObjectStorageClient{}, nil, err
	}

	return client, context.Background(), nil
}

func migrateToOCI() {
	fmt.Println("\nIniciando migración a OCI...")

	client, ctx, err := getOCIClient()
	if err != nil {
		fmt.Printf("Error inicializando cliente OCI: %v\n", err)
		return
	}

	envMap := loadEnvMap()
	namespace := envMap["OCI_NAMESPACE"]
	bucketName := envMap["OCI_BUCKET_NAME"]
	baseDir := "assets/users"

	err = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath, _ := filepath.Rel(baseDir, path)
			// Convert backslashes to slashes for object storage keys
			objectName := filepath.ToSlash(relPath)

			fmt.Printf("Subiendo: %s -> %s\n", path, objectName)

			file, err := os.Open(path)
			if err != nil {
				fmt.Printf("Error abriendo archivo %s: %v\n", path, err)
				return nil
			}
			defer file.Close()

			stat, _ := file.Stat()

			req := objectstorage.PutObjectRequest{
				NamespaceName: &namespace,
				BucketName:    &bucketName,
				ObjectName:    &objectName,
				PutObjectBody: file,
				ContentLength: common.Int64(stat.Size()),
			}

			_, err = client.PutObject(ctx, req)
			if err != nil {
				fmt.Printf("Error subiendo a OCI: %v\n", err)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error recorriendo directorios: %v\n", err)
	} else {
		fmt.Println("Migración a OCI completada.")
	}
}

func migrateFromOCI() {
	fmt.Println("\nIniciando descarga desde OCI...")

	client, ctx, err := getOCIClient()
	if err != nil {
		fmt.Printf("Error inicializando cliente OCI: %v\n", err)
		return
	}

	envMap := loadEnvMap()
	namespace := envMap["OCI_NAMESPACE"]
	bucketName := envMap["OCI_BUCKET_NAME"]
	baseDir := "assets/users"

	// List objects
	var start string
	fields := "name,size"
	for {
		req := objectstorage.ListObjectsRequest{
			NamespaceName: &namespace,
			BucketName:    &bucketName,
			Start:         &start,
			Limit:         common.Int(100),
			Fields:        &fields,
		}

		resp, err := client.ListObjects(ctx, req)
		if err != nil {
			fmt.Printf("Error listando objetos: %v\n", err)
			return
		}

		for _, item := range resp.ListObjects.Objects {
			objectName := *item.Name
			targetPath := filepath.Join(baseDir, objectName)

			fmt.Printf("Descargando: %s -> %s\n", objectName, targetPath)

			// Ensure dir
			os.MkdirAll(filepath.Dir(targetPath), 0755)

			// Get content
			getReq := objectstorage.GetObjectRequest{
				NamespaceName: &namespace,
				BucketName:    &bucketName,
				ObjectName:    &objectName,
			}

			getResp, err := client.GetObject(ctx, getReq)
			if err != nil {
				fmt.Printf("Error descargando objeto: %v\n", err)
				continue
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				fmt.Printf("Error creando archivo local: %v\n", err)
				getResp.Content.Close()
				continue
			}

			_, err = io.Copy(outFile, getResp.Content)
			outFile.Close()
			getResp.Content.Close()

			if err != nil {
				fmt.Printf("Error escribiendo archivo: %v\n", err)
			}
		}

		if resp.NextStartWith == nil {
			break
		}
		start = *resp.NextStartWith
	}

	fmt.Println("Descarga completada.")
}

func loadEnvMap() map[string]string {
	m := make(map[string]string)
	content, err := os.ReadFile("env.joss")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				// Remove quotes if present
				val = strings.Trim(val, "\"")
				m[key] = val
			}
		}
	}
	return m
}
