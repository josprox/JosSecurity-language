package main

import (
	"fmt"
	"os"

	"github.com/jossecurity/joss/pkg/core"
)

func runMigrateFresh() {
	fmt.Println("Eliminando todas las tablas y ejecutando migraciones desde cero...")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n[Error de Ejecución JOSS en Migraciones] %v\n", r)
			os.Exit(1)
		}
	}()

	// 1. Initialize Runtime
	rt := core.NewRuntime()
	rt.LoadEnv(nil)

	if rt.GetDB() == nil {
		fmt.Println("Error: No se pudo conectar a la base de datos.")
		return
	}
	fmt.Println("Conexión a DB exitosa.")

	// 2. Drop all tables
	fmt.Println("Eliminando todas las tablas...")
	rt.DropAllTables()

	// 3. Recreate migration table
	if err := rt.EnsureMigrationTable(); err != nil {
		fmt.Printf("Error creando tabla de migraciones: %v\n", err)
		os.Exit(1)
	}
	rt.EnsureAuthTables()

	// 4. Run migrations
	if err := performMigrations(rt); err != nil {
		fmt.Printf("Error ejecutando migraciones: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("¡Migraciones ejecutadas exitosamente!")
}
