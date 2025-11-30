package main

import (
	"fmt"

	"github.com/jossecurity/joss/pkg/core"
)

func runMigrateFresh() {
	fmt.Println("Eliminando todas las tablas y ejecutando migraciones desde cero...")

	// 1. Initialize Runtime
	rt := core.NewRuntime()
	rt.LoadEnv(nil)

	if rt.DB == nil {
		fmt.Println("Error: No se pudo conectar a la base de datos.")
		return
	}
	fmt.Println("Conexión a DB exitosa.")

	// 2. Drop all tables
	fmt.Println("Eliminando todas las tablas...")
	rt.DropAllTables()

	// 3. Recreate migration table
	rt.EnsureMigrationTable()
	rt.EnsureAuthTables()

	// 4. Run migrations
	performMigrations(rt)

	fmt.Println("¡Migraciones ejecutadas exitosamente!")
}
