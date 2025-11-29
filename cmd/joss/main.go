package main

import (
	"database/sql"
	"fmt"
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
	case "program":
		if len(os.Args) >= 3 && os.Args[2] == "start" {
			startProgram()
		} else {
			fmt.Println("Uso: joss program start")
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
		target := "web"
		if len(os.Args) >= 3 {
			target = os.Args[2]
		}
		if target == "program" {
			buildProgram()
		} else {
			buildWeb()
		}
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
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
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
	fmt.Println("  migrate                 - Ejecuta migraciones pendientes")
	fmt.Println("  new [web|console] [path]- Crea un nuevo proyecto")
	fmt.Println("  change db [motor]       - Cambia el motor de base de datos (mysql/sqlite)")
	fmt.Println("  version                 - Muestra la versión actual")
	fmt.Println("  help                    - Muestra esta ayuda")
}
