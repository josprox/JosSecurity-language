package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func createController(name string) {
	path := filepath.Join("app", "controllers", name+".joss")
	os.MkdirAll(filepath.Dir(path), 0755)

	content := fmt.Sprintf(`class %s {
    function index() {
        return View.render("welcome")
    }
}`, name)

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error creando controlador: %v\n", err)
		return
	}
	fmt.Printf("Controlador creado: %s\n", path)
}

func createModel(name string) {
	path := filepath.Join("app", "models", name+".joss")
	os.MkdirAll(filepath.Dir(path), 0755)

	content := fmt.Sprintf(`class %s extends GranMySQL {
    Init constructor() {
        $this->tabla = "js_%s"
    }
}`, name, strings.ToLower(name))

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error creando modelo: %v\n", err)
		return
	}
	fmt.Printf("Modelo creado: %s\n", path)
}
