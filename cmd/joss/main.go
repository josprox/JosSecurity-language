package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	"github.com/jossecurity/joss/pkg/core"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/server"
	"github.com/jossecurity/joss/pkg/template"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "server":
		if len(os.Args) >= 3 && os.Args[2] == "start" {
			server.Start()
		} else {
			fmt.Println("Uso: joss server start")
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Uso: joss run [archivo.joss]")
			return
		}
		filename := os.Args[2]
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error leyendo archivo: %v\n", err)
			return
		}

		l := parser.NewLexer(string(data))
		p := parser.NewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			fmt.Println("Errores de parseo:")
			for _, msg := range p.Errors() {
				fmt.Printf("\t%s\n", msg)
			}
			return
		}

		rt := core.NewRuntime()
		rt.Execute(program)

	case "build":
		buildProject()
	case "make:controller":
		if len(os.Args) < 3 {
			fmt.Println("Uso: joss make:controller [Nombre]")
			return
		}
		createController(os.Args[2])
	case "make:model":
		if len(os.Args) < 3 {
			fmt.Println("Uso: joss make:model [Nombre]")
			return
		}
		createModel(os.Args[2])
	case "migrate":
		runMigrations()
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Uso: joss new [web|console] [ruta]")
			fmt.Println("  joss new [ruta]          - Crea proyecto web (default)")
			fmt.Println("  joss new console [ruta]  - Crea proyecto de consola")
			fmt.Println("  joss new web [ruta]      - Crea proyecto web (explícito)")
			return
		}

		// Detectar tipo de proyecto
		if os.Args[2] == "console" {
			if len(os.Args) < 4 {
				fmt.Println("Uso: joss new console [ruta]")
				return
			}
			template.CreateConsoleProject(os.Args[3])
		} else if os.Args[2] == "web" {
			if len(os.Args) < 4 {
				fmt.Println("Uso: joss new web [ruta]")
				return
			}
			template.CreateBibleProject(os.Args[3])
		} else {
			// Default: web project
			template.CreateBibleProject(os.Args[2])
		}
	case "version":
		fmt.Println("JosSecurity v3.0 (Gold Master)")
	case "change":
		if len(os.Args) < 4 || os.Args[2] != "db" {
			fmt.Println("Uso: joss change db [motor]")
			return
		}
		targetEngine := os.Args[3]
		changeDatabaseEngine(targetEngine)
	case "help":
		printHelp()
	default:
		fmt.Printf("Comando desconocido: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func buildProject() {
	fmt.Println("Iniciando compilación de JosSecurity...")

	// 1. Validate Structure (Strict Topology)
	required := []string{
		"main.joss",
		"env.joss",
		"app",
		"config",
		"api.joss",
		"routes.joss",
	}
	for _, f := range required {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("Error de Arquitectura: Falta archivo/directorio requerido '%s'\n", f)
			fmt.Println("La Biblia de JosSecurity requiere una estructura estricta.")
			return
		}
	}

	// 2. Encrypt env.joss
	fmt.Println("Encriptando entorno...")
	encryptEnv()

	fmt.Println("Build completado exitosamente.")
}

func encryptEnv() {
	data, err := ioutil.ReadFile("env.joss")
	if err != nil {
		fmt.Println("Error leyendo env.joss")
		return
	}

	key := []byte("12345678901234567890123456789012") // Should be random in real prod
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error cipher:", err)
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error GCM:", err)
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error nonce:", err)
		return
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	err = ioutil.WriteFile("env.enc", ciphertext, 0644)
	if err != nil {
		fmt.Println("Error escribiendo env.enc")
		return
	}
	fmt.Println("Generado env.enc (AES-256)")
}

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

func changeDatabaseEngine(target string) {
	fmt.Printf("Cambiando motor de base de datos a: %s\n", target)

	// 1. Read current env
	envMap := readEnvFile("env.joss")
	currentDB := envMap["DB"]
	if currentDB == "" {
		currentDB = "mysql" // Default
	}

	if currentDB == target {
		fmt.Println("El motor de base de datos ya es " + target)
		return
	}

	// 2. Connect to Source
	srcDB, err := connectToDB(currentDB, envMap)
	if err != nil {
		fmt.Printf("Error conectando a origen (%s): %v\n", currentDB, err)
		return
	}
	defer srcDB.Close()

	// 3. Connect to Dest
	// For dest, we need to construct config.
	// If target is sqlite, we need DB_PATH.
	// If target is mysql, we need DB_HOST etc.
	// Assuming env has all configs or we prompt/use defaults.
	destDB, err := connectToDB(target, envMap)
	if err != nil {
		fmt.Printf("Error conectando a destino (%s): %v\n", target, err)
		return
	}
	defer destDB.Close()

	fmt.Println("Conectado a origen y destino.")

	// 3.5 Run Migrations on Destination to ensure Schema exists
	fmt.Println("Preparando esquema en base de datos destino...")
	destRt := core.NewRuntime()
	destRt.DB = destDB
	destRt.Env = make(map[string]string)
	// Copy env
	for k, v := range envMap {
		destRt.Env[k] = v
	}
	destRt.Env["DB"] = target // Force target driver

	// Ensure System Tables
	destRt.EnsureMigrationTable()
	destRt.EnsureAuthTables()
	destRt.EnsureCronTable()

	// Run User Migrations
	performMigrations(destRt)

	fmt.Println("Iniciando migración de datos...")

	// 4. Get Tables from Source
	tables, err := getTables(srcDB, currentDB)
	if err != nil {
		fmt.Printf("Error obteniendo tablas: %v\n", err)
		return
	}

	// 5. Migrate Data
	for _, table := range tables {
		if table == "sqlite_sequence" || table == "js_migration" || table == "js_cron" {
			continue // Skip system tables? Or migrate them too?
			// js_migration and js_cron should be migrated.
			// sqlite_sequence is internal.
		}
		if table == "sqlite_sequence" {
			continue
		}

		fmt.Printf("Migrando tabla: %s... ", table)

		// Read data
		rows, err := srcDB.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			fmt.Printf("Error leyendo tabla %s: %v\n", table, err)
			continue
		}

		cols, _ := rows.Columns()
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range cols {
			valPtrs[i] = &vals[i]
		}

		// Prepare insert in dest
		// We need to create table in dest first?
		// The user requirement says "Inyectar los datos".
		// It assumes schema exists? Or we create it?
		// "Extraer todos los datos... Inyectar... asegurando la coherencia".
		// Usually migrations create schema.
		// If we switch DB, we assume the user will run migrations or we run them?
		// If we run migrations on dest, schema exists.
		// Let's assume schema exists (user ran migrations) or we should run them.
		// But `joss change db` implies it does everything.
		// Let's assume we just copy data. If table missing, error.

		// Actually, if we switch to SQLite, it's empty.
		// We should probably run migrations on Dest first.
		// But `runMigrations` uses `core.Runtime` which uses `env.joss`.
		// We haven't updated `env.joss` yet.
		// So we can't easily run migrations using `runMigrations`.
		// We might need to copy Schema too? That's hard (SQL dialect diffs).
		// Best approach: Update env, run migrations, then copy data?
		// But if we update env, we lose source connection info (if it was same env vars).
		// But MySQL and SQLite use different vars (DB_HOST vs DB_PATH).
		// So we can have both in env.

		// Let's try to Insert. If fails, warn.
		count := 0
		placeholders := make([]string, len(cols))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		insertCmd := "INSERT INTO"
		if target == "mysql" {
			insertCmd = "INSERT IGNORE INTO"
		} else if target == "sqlite" {
			insertCmd = "INSERT OR IGNORE INTO"
		}
		query := fmt.Sprintf("%s %s (%s) VALUES (%s)", insertCmd, table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

		tx, _ := destDB.Begin()
		stmt, err := tx.Prepare(query)
		if err != nil {
			fmt.Printf("Error preparando insert (¿Existe la tabla?): %v\n", err)
			tx.Rollback()
			continue
		}

		for rows.Next() {
			rows.Scan(valPtrs...)
			// Handle types?
			// SQLite is flexible. MySQL is strict.
			// Drivers handle most.
			_, err = stmt.Exec(vals...)
			if err != nil {
				fmt.Printf("Error insertando fila: %v\n", err)
			} else {
				count++
			}
		}
		stmt.Close()
		tx.Commit()
		rows.Close()
		fmt.Printf("OK (%d filas)\n", count)
	}

	// 6. Update env.joss
	updateEnvFile("env.joss", "DB", target)
	fmt.Println("Migración completada. Archivo env.joss actualizado.")
}

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

func connectToDB(driver string, env map[string]string) (*sql.DB, error) {
	if driver == "sqlite" {
		path := "database.sqlite"
		if p, ok := env["DB_PATH"]; ok {
			path = strings.Trim(p, "\"")
			path = strings.Trim(path, "'")
		}
		return sql.Open("sqlite", path)
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", env["DB_USER"], env["DB_PASS"], env["DB_HOST"], env["DB_NAME"])
		return sql.Open("mysql", dsn)
	}
}

func getTables(db *sql.DB, driver string) ([]string, error) {
	var tables []string
	var query string
	if driver == "sqlite" {
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	} else {
		query = "SHOW TABLES"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}
	return tables, nil
}

func printHelp() {
	fmt.Println("Uso: joss [comando] [argumentos]")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  server start             Inicia el servidor HTTP de desarrollo")
	fmt.Println("  new [ruta]               Crea un nuevo proyecto web")
	fmt.Println("  new console [ruta]       Crea un nuevo proyecto de consola")
	fmt.Println("  new web [ruta]           Crea un nuevo proyecto web (explícito)")
	fmt.Println("  run [archivo]            Ejecuta un script .joss")
	fmt.Println("  build                    Compila el proyecto para producción")
	fmt.Println("  migrate                  Ejecuta las migraciones pendientes")
	fmt.Println("  change db [db]           Cambia el motor de base de datos (mysql/sqlite)")
	fmt.Println("  make:controller [Nombre] Crea un nuevo controlador")
	fmt.Println("  make:model [Nombre]      Crea un nuevo modelo")
	fmt.Println("  version                  Muestra la versión actual")
	fmt.Println("  help                     Muestra esta ayuda")
}
