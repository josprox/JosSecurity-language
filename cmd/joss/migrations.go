package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
)

func runMigrations() {
	fmt.Println("Ejecutando migraciones...")

	// 1. Initialize Runtime
	rt := core.NewRuntime()
	rt.LoadEnv()

	if rt.DB == nil {
		fmt.Println("Error: No se pudo conectar a la base de datos.")
		return
	}
	fmt.Println("Conexión a DB exitosa.")

	// Ensure migration table exists
	rt.EnsureMigrationTable()
	rt.EnsureAuthTables()

	performMigrations(rt)
}

func performMigrations(rt *core.Runtime) {
	// 2. Find migration files
	files, err := filepath.Glob("app/database/migrations/*.joss")
	if err != nil {
		fmt.Printf("Error buscando migraciones: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No se encontraron migraciones en app/database/migrations/")
		return
	}

	// 3. Get executed migrations
	executed := rt.GetExecutedMigrations()
	batch := rt.GetNextBatch()
	count := 0

	// 4. Execute pending migrations
	for _, file := range files {
		filename := filepath.Base(file)
		if executed[filename] {
			continue
		}

		fmt.Printf("Migrando: %s (Batch %d)...\n", filename, batch)

		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error leyendo %s: %v\n", file, err)
			continue
		}

		l := parser.NewLexer(string(data))
		p := parser.NewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			fmt.Printf("Error de parseo en %s:\n", file)
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			continue
		}

		rt.Execute(program)
		rt.LogMigration(filename, batch)
		count++
	}

	if count == 0 {
		fmt.Println("No hay migraciones pendientes.")
	} else {
		fmt.Printf("Migraciones completadas: %d\n", count)
	}
}
