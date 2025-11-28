package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
			fmt.Println("Uso: joss new [ruta]")
			return
		}
		template.CreateBibleProject(os.Args[2])
	case "version":
		fmt.Println("JosSecurity v3.0 (Gold Master)")
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

	// 3. Execute each migration
	for _, file := range files {
		fmt.Printf("Migrando: %s...\n", filepath.Base(file))

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
	}
	fmt.Println("Migraciones completadas.")
}

func printHelp() {
	fmt.Println("Uso: joss [comando] [argumentos]")
	fmt.Println("Comandos disponibles:")
	fmt.Println("  server start   Inicia el servidor HTTP de desarrollo")
	fmt.Println("  version        Muestra la versión actual")
	fmt.Println("  help           Muestra esta ayuda")
}
