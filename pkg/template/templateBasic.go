package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jossecurity/joss/pkg/template/files"
)

// CreateBibleProject crea un nuevo proyecto con la estructura de La Biblia de JosSecurity
func CreateBibleProject(path string) {
	fmt.Printf("Creando proyecto JosSecurity (Estructura Biblia) en: %s\n", path)

	// Create directory structure
	dirs := []string{
		filepath.Join(path, "config"),
		filepath.Join(path, "app", "models"),
		filepath.Join(path, "app", "controllers"),
		filepath.Join(path, "app", "views", "layouts"),
		filepath.Join(path, "app", "views", "auth"),
		filepath.Join(path, "app", "views", "dashboard"),
		filepath.Join(path, "app", "database", "migrations"),
		filepath.Join(path, "app", "libs"),
		filepath.Join(path, "assets", "css"),
		filepath.Join(path, "assets", "js"),
		filepath.Join(path, "assets", "images"),
		filepath.Join(path, "public", "css"),
		filepath.Join(path, "public", "js"),
		filepath.Join(path, "public", "images"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Error creando directorio %s: %v\n", dir, err)
			return
		}
	}

	// Collect all files from modules
	allFiles := make(map[string]string)

	// Helper to merge maps
	merge := func(source map[string]string) {
		for k, v := range source {
			allFiles[k] = v
		}
	}

	merge(files.GetConfigFiles(path))
	merge(files.GetRoutesFiles(path))
	merge(files.GetControllerFiles(path))
	merge(files.GetModelFiles(path))
	merge(files.GetViewFiles(path))
	merge(files.GetAssetFiles(path))

	// Write files
	for file, content := range allFiles {
		err := ioutil.WriteFile(file, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Error creando archivo %s: %v\n", file, err)
		}
	}

	fmt.Println("\n✓ Proyecto creado exitosamente")
	fmt.Printf("  cd %s\n", path)
	fmt.Println("  joss server start")
}
