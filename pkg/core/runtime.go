package core

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jossecurity/joss/pkg/parser"
	"github.com/jossecurity/joss/pkg/version"
	_ "modernc.org/sqlite"
)

var (
	// BroadcastFunc is a hook for WebSocket broadcasting
	BroadcastFunc func(msg interface{})

	runtimePool = sync.Pool{
		New: func() interface{} {
			r := &Runtime{
				Env:               make(map[string]string),
				Variables:         make(map[string]interface{}),
				VarTypes:          make(map[string]string),
				Classes:           make(map[string]*parser.ClassStatement),
				Functions:         make(map[string]*parser.MethodStatement),
				Routes:            make(map[string]map[string]interface{}),
				CurrentMiddleware: make([]string, 0),
			}
			r.Variables["cout"] = &Cout{}
			r.Variables["cin"] = &Cin{}
			r.Variables["JOSS_VERSION"] = version.Version
			r.RegisterNativeClasses()
			return r
		},
	}
)

// NewRuntime gets a runtime from the pool
func NewRuntime() *Runtime {
	r := runtimePool.Get().(*Runtime)
	// Ensure native classes are registered (if recycled)
	if _, ok := r.Variables["View"]; !ok {
		r.Variables["cout"] = &Cout{}
		r.Variables["cin"] = &Cin{}
		r.Variables["JOSS_VERSION"] = version.Version
		r.RegisterNativeClasses()
	}
	return r
}

// FreeRuntime returns the runtime to the pool
func (r *Runtime) Free() {
	// Reset state
	for k := range r.Variables {
		delete(r.Variables, k)
	}
	// Restore standard variables
	r.Variables["cout"] = &Cout{}
	r.Variables["cin"] = &Cin{}

	// Keep Env, Classes, Functions, Routes as they are likely static or re-loaded?
	// If Routes are dynamic per request (e.g. defined in routes.joss which is parsed every time?), then we should clear them.
	// But parsing every time is slow.
	// For now, let's assume we clear Variables.
	// We should also clear CurrentMiddleware
	r.CurrentMiddleware = r.CurrentMiddleware[:0]

	runtimePool.Put(r)
}

// LoadEnv loads environment variables from env.joss
func (r *Runtime) LoadEnv() {
	fmt.Println("[Security] Cargando entorno...")

	// Try reading env.joss
	content, err := os.ReadFile("env.joss")
	if err != nil {
		// Try looking in parent directory (for examples/ scripts)
		content, err = os.ReadFile("../env.joss")
		if err != nil {
			// Try looking in project root if running from subfolder
			content, err = os.ReadFile("../../env.joss")
			if err != nil {
				// Try looking in the specific test folder
				content, err = os.ReadFile("pruebas 271125/env.joss")
				if err != nil {
					fmt.Println("[Security] Advertencia: No se encontró env.joss")
					return
				}
			}
		}
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Remove quotes if present
			val = strings.Trim(val, "\"")
			val = strings.Trim(val, "'")
			r.Env[key] = val
		}
	}

	// Connect to DB
	dbDriver := "mysql"
	if val, ok := r.Env["DB"]; ok {
		dbDriver = val
	}

	var dsn string

	if dbDriver == "sqlite" {
		dbPath := "database.sqlite"
		if val, ok := r.Env["DB_PATH"]; ok {
			dbPath = val
		}
		dsn = dbPath
		fmt.Printf("[Security] Conectando a SQLite: %s\n", dbPath)
		// Ensure sqlite3 driver is imported
	} else {
		// Default to MySQL
		if host, ok := r.Env["DB_HOST"]; ok {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", r.Env["DB_USER"], r.Env["DB_PASS"], host, r.Env["DB_NAME"])
			fmt.Printf("[Security] Conectando a MySQL: %s\n", host)
		} else {
			// No DB config found
			return
		}
	}

	db, err := sql.Open(dbDriver, dsn)
	if err == nil {
		// db.Ping() // Optional: don't block if DB is down
		r.DB = db
		r.EnsureCronTable()
		r.EnsureMigrationTable()
		r.EnsureAuthTables()
	} else {
		fmt.Printf("[Security] Error conectando a DB: %v\n", err)
	}
}

// NewInstance creates a new instance of a class
func NewInstance(class *parser.ClassStatement) *Instance {
	return &Instance{
		Class:  class,
		Fields: make(map[string]interface{}),
	}
}
