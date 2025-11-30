package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func readEnvFile(path string) map[string]string {
	m := make(map[string]string)
	content, _ := ioutil.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"")
			val = strings.Trim(val, "'")
			m[strings.TrimSpace(parts[0])] = val
		}
	}
	return m
}

func updateEnvFile(path, key, value string) {
	content, _ := ioutil.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	found := false
	var newLines []string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}
	if !found {
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}
	ioutil.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func printHelp() {
	fmt.Println("Uso: joss [comando] [argumentos]")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  server start            - Inicia el servidor web")
	fmt.Println("  program start           - Inicia la aplicación en modo escritorio")
	fmt.Println("  run [archivo]           - Ejecuta un script .joss")
	fmt.Println("  build [web|program]     - Compila el proyecto para distribución")
	fmt.Println("  make:controller [Name]  - Crea un nuevo controlador")
	fmt.Println("  make:model [Name]       - Crea un nuevo modelo")
	fmt.Println("  make:view [Name]        - Crea una nueva vista (solo web)")
	fmt.Println("  make:mvc [Name]         - Crea Modelo, Vista y Controlador")
	fmt.Println("  make:crud [Tabla]       - Crea CRUD completo basado en tabla")
	fmt.Println("  make:migration [Name]   - Crea una nueva migración")
	fmt.Println("  migrate                 - Ejecuta migraciones pendientes")
	fmt.Println("  migrate:fresh           - Elimina todas las tablas y re-ejecuta migraciones")
	fmt.Println("  new [web|console] [path]- Crea un nuevo proyecto")
	fmt.Println("  change db [motor]       - Cambia el motor de base de datos (mysql/sqlite)")
	fmt.Println("  version                 - Muestra la versión actual")
	fmt.Println("  help                    - Muestra esta ayuda")
}
